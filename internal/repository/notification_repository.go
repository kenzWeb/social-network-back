package repository

import (
	"context"
	"modern-social-media/internal/models"

	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func (r NotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r NotificationRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.WithContext(ctx).
		Preload("Actor").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

func (r NotificationRepository) GetUnreadByUserID(ctx context.Context, userID string, limit, offset int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.WithContext(ctx).
		Preload("Actor").
		Where("user_id = ? AND read = ?", userID, false).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

func (r NotificationRepository) CountUnread(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Count(&count).Error
	return count, err
}

func (r NotificationRepository) MarkAsRead(ctx context.Context, notificationID, userID string) error {
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("read", true).Error
}

func (r NotificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Update("read", true).Error
}

func (r NotificationRepository) Delete(ctx context.Context, notificationID, userID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Delete(&models.Notification{}).Error
}

func (r NotificationRepository) DeleteByActorAndType(ctx context.Context, userID, actorID string, notifType models.NotificationType) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND actor_id = ? AND type = ?", userID, actorID, notifType).
		Delete(&models.Notification{}).Error
}

func (r NotificationRepository) Exists(ctx context.Context, userID, actorID string, notifType models.NotificationType, targetID *string) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND actor_id = ? AND type = ?", userID, actorID, notifType)

	if targetID != nil {
		query = query.Where("target_id = ?", *targetID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}
