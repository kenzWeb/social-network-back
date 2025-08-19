package repository

import (
	"context"
	"modern-social-media/internal/models"

	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func (r PostRepository) GetById(ctx context.Context, id string) (*models.Post, error) {
	var post models.Post

	if err := r.db.WithContext(ctx).First(&id, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &post, nil
}

func (r PostRepository) GetAllPosts(ctx context.Context) ([]models.Post, error) {
	var posts []models.Post

	if err := r.db.WithContext(ctx).Find(&posts).Error; err != nil {
		return nil, err
	}

	return posts, nil
}

func (r PostRepository) CreatePost(ctx context.Context, p *models.Post) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r PostRepository) UpdatePost(ctx context.Context, p *models.Post) error {
	return r.db.WithContext(ctx).Create(p).Error
}
