package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Store struct {
	client *goredis.Client
}

func New(client *goredis.Client) *Store {
	return &Store{client: client}
}

func (s *Store) Blacklist(ctx context.Context, token string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = time.Minute
	}
	return s.client.Set(ctx, tokenKey(token), "1", ttl).Err()
}

func (s *Store) Contains(ctx context.Context, token string) (bool, error) {
	ok, err := s.client.Exists(ctx, tokenKey(token)).Result()
	if err != nil {
		return false, err
	}
	return ok > 0, nil
}

func tokenKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return "blacklist:token:" + hex.EncodeToString(sum[:])
}
