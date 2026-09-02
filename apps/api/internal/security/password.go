// Package security holds password hashing.
//
// This exists because the legacy `user` table stores passwords in plaintext and
// the old login compared them literally in SQL. Nothing about that survives: no
// credential is migrated, and everything from here on is hashed.
//
// The driver is selectable, as in Laravel:
//
//	HASH_DRIVER=bcrypt   (default, Laravel's default too)
//	HASH_DRIVER=argon2id
//	BCRYPT_ROUNDS=12
//
// Verify detects the algorithm from the stored hash rather than from config, so
// changing the driver does not lock anyone out — existing hashes keep verifying
// and NeedsRehash flags them for transparent upgrade on next login.
package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

type Driver string

const (
	Bcrypt   Driver = "bcrypt"
	Argon2id Driver = "argon2id"
)

// argon2id parameters follow the OWASP guidance. They are stored inside the
// hash, so they can be raised later without invalidating existing passwords.
const (
	defaultMemory  = 64 * 1024 // 64 MiB
	defaultTime    = 3
	defaultKeyLen  = 32
	defaultSaltLen = 16

	// bcrypt's own maximum. Anything past 72 bytes is silently ignored by the
	// algorithm, which would mean two different long passwords sharing a hash.
	bcryptMaxPasswordBytes = 72
	defaultBcryptCost      = 12
)

var (
	ErrInvalidHash = errors.New("security: unrecognised password hash format")
	ErrMismatch    = errors.New("security: password does not match")
	ErrTooLong     = fmt.Errorf("security: password exceeds %d bytes, which bcrypt truncates", bcryptMaxPasswordBytes)
)

// CurrentDriver reports the configured driver, defaulting to bcrypt.
func CurrentDriver() Driver {
	switch strings.ToLower(os.Getenv("HASH_DRIVER")) {
	case string(Argon2id), "argon2", "argon":
		return Argon2id
	default:
		return Bcrypt
	}
}

func bcryptCost() int {
	if v := os.Getenv("BCRYPT_ROUNDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= bcrypt.MinCost && n <= bcrypt.MaxCost {
			return n
		}
	}
	return defaultBcryptCost
}

// Hash hashes a password with the configured driver.
func Hash(password string) (string, error) {
	if password == "" {
		return "", errors.New("security: password is empty")
	}
	if CurrentDriver() == Argon2id {
		return hashArgon2id(password)
	}
	return hashBcrypt(password)
}

// Verify reports whether password matches encoded, choosing the algorithm from
// the hash itself.
func Verify(password, encoded string) error {
	switch {
	case strings.HasPrefix(encoded, "$argon2id$"):
		return verifyArgon2id(password, encoded)
	case strings.HasPrefix(encoded, "$2a$"),
		strings.HasPrefix(encoded, "$2b$"),
		strings.HasPrefix(encoded, "$2y$"):
		return verifyBcrypt(password, encoded)
	default:
		// Anything else — including a plaintext value carried over from the
		// legacy table — is refused outright rather than compared.
		return ErrInvalidHash
	}
}

// NeedsRehash reports whether a stored hash should be upgraded on next login,
// either because it uses a different driver than the current one or because its
// parameters are weaker than the current defaults.
func NeedsRehash(encoded string) bool {
	switch {
	case strings.HasPrefix(encoded, "$argon2id$"):
		if CurrentDriver() != Argon2id {
			return true
		}
		p, _, _, err := decodeArgon2id(encoded)
		if err != nil {
			return true
		}
		d := argonDefaults()
		return p.memory < d.memory || p.time < d.time || p.keyLen < d.keyLen
	case strings.HasPrefix(encoded, "$2a$"),
		strings.HasPrefix(encoded, "$2b$"),
		strings.HasPrefix(encoded, "$2y$"):
		if CurrentDriver() != Bcrypt {
			return true
		}
		cost, err := bcrypt.Cost([]byte(encoded))
		return err != nil || cost < bcryptCost()
	default:
		return true
	}
}

// ---------------------------------------------------------------- bcrypt ---

func hashBcrypt(password string) (string, error) {
	// bcrypt ignores everything past 72 bytes. Silently truncating means two
	// different long passwords can produce the same hash, so refuse instead.
	if len(password) > bcryptMaxPasswordBytes {
		return "", ErrTooLong
	}
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost())
	if err != nil {
		return "", fmt.Errorf("security: bcrypt: %w", err)
	}
	return string(h), nil
}

func verifyBcrypt(password, encoded string) error {
	err := bcrypt.CompareHashAndPassword([]byte(encoded), []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrMismatch
	}
	if err != nil {
		return ErrInvalidHash
	}
	return nil
}

// -------------------------------------------------------------- argon2id ---

type argonParams struct {
	memory  uint32
	time    uint32
	threads uint8
	keyLen  uint32
}

func argonDefaults() argonParams {
	threads := uint8(runtime.NumCPU())
	if threads > 4 {
		threads = 4
	}
	if threads == 0 {
		threads = 1
	}
	return argonParams{memory: defaultMemory, time: defaultTime, threads: threads, keyLen: defaultKeyLen}
}

// hashArgon2id returns a PHC-format string:
//
//	$argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
//
// The parameters travel with the hash, which is what makes them changeable
// without invalidating every existing password.
func hashArgon2id(password string) (string, error) {
	p := argonDefaults()
	salt := make([]byte, defaultSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("security: read salt: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, p.time, p.memory, p.threads, p.keyLen)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.memory, p.time, p.threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// verifyArgon2id compares in constant time: a timing-variable compare on a
// password hash leaks how much of a guess was correct.
func verifyArgon2id(password, encoded string) error {
	p, salt, want, err := decodeArgon2id(encoded)
	if err != nil {
		return err
	}
	got := argon2.IDKey([]byte(password), salt, p.time, p.memory, p.threads, p.keyLen)
	if subtle.ConstantTimeCompare(got, want) != 1 {
		return ErrMismatch
	}
	return nil
}

func decodeArgon2id(encoded string) (argonParams, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return argonParams{}, nil, nil, ErrInvalidHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return argonParams{}, nil, nil, ErrInvalidHash
	}
	if version != argon2.Version {
		return argonParams{}, nil, nil, fmt.Errorf("%w: unsupported version %d", ErrInvalidHash, version)
	}

	var p argonParams
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.memory, &p.time, &p.threads); err != nil {
		return argonParams{}, nil, nil, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil {
		return argonParams{}, nil, nil, ErrInvalidHash
	}
	key, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil {
		return argonParams{}, nil, nil, ErrInvalidHash
	}
	p.keyLen = uint32(len(key))

	return p, salt, key, nil
}
