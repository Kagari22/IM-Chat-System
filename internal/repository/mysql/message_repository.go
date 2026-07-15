package mysql

import (
	"context"
	"database/sql"

	"IM_Chat_System/internal/model"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Save(ctx context.Context, message model.Message) (model.Message, error) {
	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO messages (from_user_id, to_user_id, content_type, content, object_key, object_url, file_name, file_size) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		message.FromUserID,
		message.ToUserID,
		message.ContentType,
		message.Content,
		message.ObjectKey,
		message.ObjectURL,
		message.FileName,
		message.FileSize,
	)
	if err != nil {
		return model.Message{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Message{}, err
	}

	saved := model.Message{}
	err = r.db.QueryRowContext(
		ctx,
		`SELECT id, from_user_id, to_user_id, content_type, content, object_key, object_url, file_name, file_size, created_at FROM messages WHERE id = ?`,
		id,
	).Scan(&saved.ID, &saved.FromUserID, &saved.ToUserID, &saved.ContentType, &saved.Content, &saved.ObjectKey, &saved.ObjectURL, &saved.FileName, &saved.FileSize, &saved.CreatedAt)
	if err != nil {
		return model.Message{}, err
	}
	return saved, nil
}

func (r *MessageRepository) ListConversation(ctx context.Context, userID, peerID, afterID int64, limit int) ([]model.Message, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, from_user_id, to_user_id, content_type, content, object_key, object_url, file_name, file_size, created_at
		 FROM messages
		 WHERE id > ?
		   AND ((from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?))
		 ORDER BY id ASC
		 LIMIT ?`,
		afterID, userID, peerID, peerID, userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var message model.Message
		if err := rows.Scan(&message.ID, &message.FromUserID, &message.ToUserID, &message.ContentType, &message.Content, &message.ObjectKey, &message.ObjectURL, &message.FileName, &message.FileSize, &message.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

func (r *MessageRepository) ListOffline(ctx context.Context, userID, afterID int64, limit int) ([]model.Message, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, from_user_id, to_user_id, content_type, content, object_key, object_url, file_name, file_size, created_at
		 FROM messages
		 WHERE to_user_id = ? AND id > ?
		 ORDER BY id ASC
		 LIMIT ?`,
		userID, afterID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var message model.Message
		if err := rows.Scan(&message.ID, &message.FromUserID, &message.ToUserID, &message.ContentType, &message.Content, &message.ObjectKey, &message.ObjectURL, &message.FileName, &message.FileSize, &message.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}
