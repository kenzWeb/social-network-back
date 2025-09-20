package repository

import (
	"context"
	"modern-social-media/internal/models"

	"gorm.io/gorm"
)

type SkillRepository struct {
	db *gorm.DB
}

func (r SkillRepository) GetAllSkills(ctx context.Context, userID string) ([]models.Skill, error) {
	var skills []models.Skill
	if err := r.db.WithContext(ctx).Model(&models.Skill{}).Where("user_id = ?", userID).Find(&skills).Error; err != nil {
		return nil, err
	}
	return skills, nil
}
