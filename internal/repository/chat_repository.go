package repository

import (
	"context"
	"errors"
	"time"

	"modern-social-media/internal/models"

	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return ChatRepository{db: db}
}

func (r ChatRepository) GetOrCreateDirectConversation(ctx context.Context, userA, userB string) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.WithContext(ctx).
		Raw(`
                SELECT c.* FROM conversations c
                WHERE EXISTS (SELECT 1 FROM conversation_participants p1 WHERE p1.conversation_id = c.id AND p1.user_id = ?)
                    AND EXISTS (SELECT 1 FROM conversation_participants p2 WHERE p2.conversation_id = c.id AND p2.user_id = ?)
                ORDER BY c.updated_at DESC LIMIT 1
                `, userA, userB).Scan(&conv).Error
	if err == nil && conv.ID != "" {
		return &conv, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	conv = models.Conversation{}
	if err := r.db.WithContext(ctx).Create(&conv).Error; err != nil {
		return nil, err
	}
	parts := []models.ConversationParticipant{
		{ConversationID: conv.ID, UserID: userA},
		{ConversationID: conv.ID, UserID: userB},
	}
	if err := r.db.WithContext(ctx).Create(&parts).Error; err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r ChatRepository) CreateMessage(ctx context.Context, msg *models.Message) error {
	if err := r.db.WithContext(ctx).Create(msg).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&models.Conversation{}).Where("id = ?", msg.ConversationID).Update("updated_at", time.Now()).Error
}

func (r ChatRepository) ListUserConversations(ctx context.Context, userID string, limit, offset int) ([]models.Conversation, error) {
	var items []models.Conversation
	err := r.db.WithContext(ctx).
		Joins("JOIN conversation_participants cp ON cp.conversation_id = conversations.id AND cp.user_id = ?", userID).
		Order("updated_at DESC").
		Limit(limit).Offset(offset).
		Find(&items).Error
	return items, err
}

func (r ChatRepository) ListMessages(ctx context.Context, conversationID string, limit, offset int) ([]models.Message, error) {
	var msgs []models.Message
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&msgs).Error
	return msgs, err
}

func (r ChatRepository) UpdateLastRead(ctx context.Context, conversationID, userID string, t time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.ConversationParticipant{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("last_read_at", t).Error
}
