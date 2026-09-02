package security

import (
	"errors"
	"strings"
	"testing"
)

func TestBcryptIsTheDefaultDriver(t *testing.T) {
	t.Setenv("HASH_DRIVER", "")
	if got := CurrentDriver(); got != Bcrypt {
		t.Errorf("default driver = %q, want bcrypt (Laravel's default too)", got)
	}
	h, err := Hash("secret")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(h, "$2") {
		t.Errorf("expected a bcrypt hash, got %q", h)
	}
}

func TestRoundTripOnBothDrivers(t *testing.T) {
	for _, driver := range []string{"bcrypt", "argon2id"} {
		t.Run(driver, func(t *testing.T) {
			t.Setenv("HASH_DRIVER", driver)
			h, err := Hash("correct horse battery staple")
			if err != nil {
				t.Fatal(err)
			}
			if err := Verify("correct horse battery staple", h); err != nil {
				t.Errorf("correct password rejected: %v", err)
			}
			if err := Verify("wrong", h); !errors.Is(err, ErrMismatch) {
				t.Errorf("wrong password accepted, got %v", err)
			}
		})
	}
}

// Switching the driver must not lock anyone out: the algorithm is read from the
// stored hash, never from configuration.
func TestAHashVerifiesAfterTheDriverChanges(t *testing.T) {
	t.Setenv("HASH_DRIVER", "argon2id")
	argonHash, err := Hash("shared")
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("HASH_DRIVER", "bcrypt")
	if err := Verify("shared", argonHash); err != nil {
		t.Errorf("argon2id hash stopped verifying after switching to bcrypt: %v", err)
	}
	if !NeedsRehash(argonHash) {
		t.Error("a hash from the non-current driver should be flagged for upgrade")
	}
}

func TestHashDoesNotContainThePassword(t *testing.T) {
	for _, driver := range []string{"bcrypt", "argon2id"} {
		t.Setenv("HASH_DRIVER", driver)
		const pw = "sup3rs3cret"
		h, err := Hash(pw)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(h, pw) {
			t.Fatalf("%s: password present in the stored hash: %q", driver, h)
		}
	}
}

// A per-password salt is what stops two people with the same password sharing a
// hash, and what makes a precomputed table useless.
func TestSamePasswordHashesDifferently(t *testing.T) {
	for _, driver := range []string{"bcrypt", "argon2id"} {
		t.Setenv("HASH_DRIVER", driver)
		a, _ := Hash("same")
		b, _ := Hash("same")
		if a == b {
			t.Errorf("%s: hashes identical, so the salt is not random", driver)
		}
	}
}

// bcrypt silently ignores everything past 72 bytes, which would let two
// different long passwords share a hash. Refusing is safer than truncating.
func TestBcryptRefusesOverlongPasswordsRatherThanTruncating(t *testing.T) {
	t.Setenv("HASH_DRIVER", "bcrypt")
	long := strings.Repeat("a", 73)
	if _, err := Hash(long); !errors.Is(err, ErrTooLong) {
		t.Fatalf("got %v, want ErrTooLong", err)
	}
	// argon2id has no such limit.
	t.Setenv("HASH_DRIVER", "argon2id")
	if _, err := Hash(long); err != nil {
		t.Errorf("argon2id should accept a long password: %v", err)
	}
}

func TestPlaintextIsRefusedNotCompared(t *testing.T) {
	// The legacy table stored passwords in plaintext. If such a value ever
	// reaches this code it must be rejected, never matched.
	if err := Verify("admin123", "admin123"); !errors.Is(err, ErrInvalidHash) {
		t.Errorf("a plaintext 'hash' was not refused: %v", err)
	}
}

func TestEmptyPasswordIsRejected(t *testing.T) {
	if _, err := Hash(""); err == nil {
		t.Error("an empty password should not be hashable")
	}
}

func TestMalformedHashesAreRejectedNotPanicking(t *testing.T) {
	for _, bad := range []string{
		"", "plaintext", "$argon2id$", "$2z$12$abcdefghijklmnopqrstuv",
		"$argon2id$v=99$m=1,t=1,p=1$c2FsdA$aGFzaA",
		"$argon2id$v=19$m=1,t=1,p=1$!!!$aGFzaA",
	} {
		if err := Verify("x", bad); err == nil {
			t.Errorf("malformed hash %q was accepted", bad)
		}
	}
}

func TestNeedsRehashFlagsWeakerParameters(t *testing.T) {
	t.Setenv("HASH_DRIVER", "bcrypt")
	t.Setenv("BCRYPT_ROUNDS", "12")
	current, _ := Hash("x")
	if NeedsRehash(current) {
		t.Error("a hash at the current cost should not need rehashing")
	}

	t.Setenv("BCRYPT_ROUNDS", "14")
	if !NeedsRehash(current) {
		t.Error("raising BCRYPT_ROUNDS should flag older hashes for upgrade")
	}
}
