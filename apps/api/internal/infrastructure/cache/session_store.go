package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"dpmptsp/api/internal/domain"
)

// SessionStore keeps sessions in Redis.
type SessionStore struct{ c *Client }

func NewSessionStore(c *Client) *SessionStore { return &SessionStore{c: c} }

var _ domain.SessionStore = (*SessionStore)(nil)

// ErrNoStore is returned when Redis is unavailable.
//
// Unlike the caches, this fails closed: without a store a session cannot be
// verified, and treating that as "no session" is the only safe reading.
var ErrNoStore = errors.New("session store unavailable")

func sessionKey(id string) string      { return "session:" + id }
func userSessionsKey(uid int64) string { return fmt.Sprintf("session:user:%d", uid) }

func (s *SessionStore) Create(ctx context.Context, sess domain.Session) error {
	if s.c == nil {
		return ErrNoStore
	}
	raw, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	ttl := time.Until(sess.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("session already expired")
	}

	pipe := s.c.rdb.TxPipeline()
	pipe.Set(ctx, sessionKey(sess.ID), raw, ttl)
	// Index by user so every session can be revoked at once.
	pipe.SAdd(ctx, userSessionsKey(sess.UserID), sess.ID)
	pipe.Expire(ctx, userSessionsKey(sess.UserID), ttl)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *SessionStore) Get(ctx context.Context, id string) (*domain.Session, error) {
	if s.c == nil {
		return nil, ErrNoStore
	}
	raw, err := s.c.rdb.Get(ctx, sessionKey(id)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read session: %w", err)
	}
	var sess domain.Session
	if err := json.Unmarshal(raw, &sess); err != nil {
		return nil, domain.ErrNotFound
	}
	if time.Now().After(sess.ExpiresAt) {
		return nil, domain.ErrNotFound
	}
	return &sess, nil
}

func (s *SessionStore) Delete(ctx context.Context, id string) error {
	if s.c == nil {
		return ErrNoStore
	}
	// Read first so the user index can be cleaned too.
	sess, err := s.Get(ctx, id)
	if err == nil {
		s.c.rdb.SRem(ctx, userSessionsKey(sess.UserID), id)
	}
	return s.c.rdb.Del(ctx, sessionKey(id)).Err()
}

func (s *SessionStore) DeleteForUser(ctx context.Context, userID int64) error {
	if s.c == nil {
		return ErrNoStore
	}
	// SMembers on a per-user set, not a SCAN over the keyspace.
	ids, err := s.c.rdb.SMembers(ctx, userSessionsKey(userID)).Result()
	if err != nil {
		return err
	}
	keys := make([]string, 0, len(ids)+1)
	for _, id := range ids {
		keys = append(keys, sessionKey(id))
	}
	keys = append(keys, userSessionsKey(userID))
	return s.c.rdb.Del(ctx, keys...).Err()
}
