package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

type Store struct {
	client *goredis.Client
}

func New(client *goredis.Client) *Store {
	return &Store{client: client}
}

func (s *Store) Increment(ctx context.Context, userID, peerID int64) (int64, error) {
	return s.client.Incr(ctx, conversationKey(userID, peerID)).Result()
}

func (s *Store) ClearConversation(ctx context.Context, userID, peerID int64) error {
	return s.client.Del(ctx, conversationKey(userID, peerID)).Err()
}

func (s *Store) GetConversationCounts(ctx context.Context, userID int64, peerIDs []int64) (map[int64]int64, error) {
	result := make(map[int64]int64, len(peerIDs))
	if len(peerIDs) == 0 {
		return result, nil
	}

	pipe := s.client.Pipeline()
	cmds := make(map[int64]*goredis.StringCmd, len(peerIDs))
	for _, peerID := range peerIDs {
		cmds[peerID] = pipe.Get(ctx, conversationKey(userID, peerID))
	}
	_, err := pipe.Exec(ctx)
	if err != nil && err != goredis.Nil {
		// Ignore Nil results from missing keys; inspect command-level errors below.
	}

	for peerID, cmd := range cmds {
		value, cmdErr := cmd.Int64()
		if cmdErr == goredis.Nil {
			result[peerID] = 0
			continue
		}
		if cmdErr != nil {
			return nil, cmdErr
		}
		result[peerID] = value
	}
	return result, nil
}

func conversationKey(userID, peerID int64) string {
	return fmt.Sprintf("unread:user:%d:peer:%d", userID, peerID)
}
