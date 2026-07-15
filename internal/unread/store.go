package unread

import "context"

type Store interface {
	Increment(ctx context.Context, userID, peerID int64) (int64, error)
	ClearConversation(ctx context.Context, userID, peerID int64) error
	GetConversationCounts(ctx context.Context, userID int64, peerIDs []int64) (map[int64]int64, error)
}
