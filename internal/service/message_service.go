package service

import (
	"context"
	"errors"
	"io"
	"log"
	"mime"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"IM_Chat_System/internal/model"
	"IM_Chat_System/internal/mq"
	"IM_Chat_System/internal/repository"
	"IM_Chat_System/internal/storage"
	"IM_Chat_System/internal/unread"
)

type MessageService struct {
	users    repository.UserRepository
	messages repository.MessageRepository
	unread   unread.Store
	events   mq.EventPublisher
	uploader storage.Uploader
}

func NewMessageService(users repository.UserRepository, messages repository.MessageRepository, unreadStore unread.Store, publisher mq.EventPublisher, uploader storage.Uploader) *MessageService {
	if publisher == nil {
		publisher = mq.NoopPublisher{}
	}
	if uploader == nil {
		uploader = storage.NoopUploader{}
	}
	return &MessageService{
		users:    users,
		messages: messages,
		unread:   unreadStore,
		events:   publisher,
		uploader: uploader,
	}
}

func (s *MessageService) ListUsers(ctx context.Context, currentUserID int64) ([]model.User, error) {
	users, err := s.users.List(ctx, currentUserID)
	if err != nil {
		return nil, err
	}
	if s.unread == nil || len(users) == 0 {
		return users, nil
	}

	peerIDs := make([]int64, 0, len(users))
	for _, user := range users {
		peerIDs = append(peerIDs, user.ID)
	}
	counts, err := s.unread.GetConversationCounts(ctx, currentUserID, peerIDs)
	if err != nil {
		return nil, err
	}
	for i := range users {
		users[i].UnreadCount = counts[users[i].ID]
	}
	return users, nil
}

func (s *MessageService) GetMe(ctx context.Context, userID int64) (model.User, error) {
	user, ok, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return model.User{}, err
	}
	if !ok {
		return model.User{}, errors.New("user not found")
	}
	user.PasswordHash = ""
	return user, nil
}

func (s *MessageService) SaveText(ctx context.Context, fromUserID, toUserID int64, content string) (model.Message, error) {
	content = strings.TrimSpace(content)
	if toUserID <= 0 || content == "" {
		return model.Message{}, errors.New("to and content are required")
	}

	if _, ok, err := s.users.GetByID(ctx, toUserID); err != nil {
		return model.Message{}, err
	} else if !ok {
		return model.Message{}, errors.New("receiver not found")
	}

	message, err := s.messages.Save(ctx, model.Message{
		FromUserID:  fromUserID,
		ToUserID:    toUserID,
		ContentType: "text",
		Content:     content,
	})
	if err != nil {
		return model.Message{}, err
	}
	if s.unread != nil {
		if _, err := s.unread.Increment(ctx, toUserID, fromUserID); err != nil {
			return model.Message{}, err
		}
	}
	if err := s.events.PublishMessageCreated(ctx, mq.NewMessageCreatedEvent(message)); err != nil {
		log.Println("publish message.created:", err)
	}
	return message, nil
}

func (s *MessageService) SaveMedia(ctx context.Context, fromUserID, toUserID int64, fileName string, size int64, contentType string, reader io.Reader) (model.Message, error) {
	if toUserID <= 0 {
		return model.Message{}, errors.New("to_user_id is required")
	}
	if _, ok, err := s.users.GetByID(ctx, toUserID); err != nil {
		return model.Message{}, err
	} else if !ok {
		return model.Message{}, errors.New("receiver not found")
	}

	fileName = filepath.Base(strings.TrimSpace(fileName))
	if fileName == "" {
		fileName = "file"
	}
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(fileName))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	objectName := buildObjectName(fromUserID, toUserID, fileName)
	object, err := s.uploader.Upload(ctx, objectName, reader, size, contentType)
	if err != nil {
		return model.Message{}, err
	}

	messageType := "file"
	if strings.HasPrefix(contentType, "image/") {
		messageType = "image"
	}

	message, err := s.messages.Save(ctx, model.Message{
		FromUserID:  fromUserID,
		ToUserID:    toUserID,
		ContentType: messageType,
		Content:     fileName,
		ObjectKey:   object.Key,
		ObjectURL:   object.URL,
		FileName:    fileName,
		FileSize:    object.Size,
	})
	if err != nil {
		return model.Message{}, err
	}
	if s.unread != nil {
		if _, err := s.unread.Increment(ctx, toUserID, fromUserID); err != nil {
			return model.Message{}, err
		}
	}
	if err := s.events.PublishMessageCreated(ctx, mq.NewMessageCreatedEvent(message)); err != nil {
		log.Println("publish message.created:", err)
	}
	return message, nil
}

func (s *MessageService) Conversation(ctx context.Context, userID, peerID, afterID int64, limit int) ([]model.Message, error) {
	if peerID <= 0 {
		return nil, errors.New("peer_id is required")
	}
	if limit <= 0 {
		limit = 50
	}
	messages, err := s.messages.ListConversation(ctx, userID, peerID, afterID, limit)
	if err != nil {
		return nil, err
	}
	if s.unread != nil {
		if err := s.unread.ClearConversation(ctx, userID, peerID); err != nil {
			return nil, err
		}
	}
	return messages, nil
}

func (s *MessageService) Offline(ctx context.Context, userID, afterID int64, limit int) ([]model.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.messages.ListOffline(ctx, userID, afterID, limit)
}

func buildObjectName(fromUserID, toUserID int64, fileName string) string {
	safeName := strings.ReplaceAll(fileName, " ", "_")
	return "chat/" + strconv.FormatInt(fromUserID, 10) + "/" + strconv.FormatInt(toUserID, 10) + "/" + strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + safeName
}
