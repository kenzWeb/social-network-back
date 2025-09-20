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
	if err := r.db.WithContext(ctx).Model(&models.Skill{}).
		Preload("User").
		Where("user_id = ?", userID).Find(&skills).Error; err != nil {
		return nil, err
	}
	return skills, nil
}

func (r SkillRepository) AddSkill(ctx context.Context, skill *models.Skill) error {
	return r.db.WithContext(ctx).Create(skill).Error
}
