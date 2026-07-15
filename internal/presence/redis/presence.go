package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const onlineTTL = 60 * time.Second

type PresenceStore struct {
	client *goredis.Client
}

func New(client *goredis.Client) *PresenceStore {
	return &PresenceStore{client: client}
}

func (s *PresenceStore) SetOnline(ctx context.Context, userID int64, node string) error {
	return s.client.Set(ctx, onlineKey(userID), node, onlineTTL).Err()
}

func (s *PresenceStore) GetOnlineNode(ctx context.Context, userID int64) (string, bool, error) {
	value, err := s.client.Get(ctx, onlineKey(userID)).Result()
	if err == goredis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

func (s *PresenceStore) SetOffline(ctx context.Context, userID int64) error {
	return s.client.Del(ctx, onlineKey(userID)).Err()
}

func onlineKey(userID int64) string {
	return fmt.Sprintf("online:user:%d", userID)
}
