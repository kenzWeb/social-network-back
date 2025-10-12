package repository

import (
	"context"
	"time"

	"modern-social-media/internal/models"

	"gorm.io/gorm"
)

type VerificationCodeRepository struct {
	db *gorm.DB
}

func (r VerificationCodeRepository) Create(ctx context.Context, v *models.VerificationCode) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r VerificationCodeRepository) GetValid(ctx context.Context, userID, purpose, code string) (*models.VerificationCode, error) {
	var v models.VerificationCode
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND purpose = ? AND code = ? AND expires_at > NOW() AND consumed_at IS NULL", userID, purpose, code).
		First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r VerificationCodeRepository) Consume(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.VerificationCode{}).
		Where("id = ? AND consumed_at IS NULL", id).
		Updates(map[string]interface{}{"consumed_at": &now}).Error
}

func (r VerificationCodeRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at <= NOW() OR consumed_at IS NOT NULL").Delete(&models.VerificationCode{}).Error
}

func (r VerificationCodeRepository) DeleteByUserAndPurpose(ctx context.Context, userID, purpose string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND purpose = ?", userID, purpose).
		Delete(&models.VerificationCode{}).Error
}
