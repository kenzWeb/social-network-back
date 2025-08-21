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
	if err := r.db.WithContext(ctx).First(&post, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (r PostRepository) GetPostsByUser(ctx context.Context, userID string) ([]models.Post, error) {
	var posts []models.Post
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r PostRepository) CreatePost(ctx context.Context, p *models.Post) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r PostRepository) UpdatePostByUser(ctx context.Context, postID, userID string, p *models.Post) error {
	var post models.Post
	if err := r.db.WithContext(ctx).First(&post, "id = ? AND user_id = ?", postID, userID).Error; err != nil {
		return err
	}
	post.Content = p.Content
	post.ImageURL = p.ImageURL

	return r.db.WithContext(ctx).Save(&post).Error
}

func (r PostRepository) DeletePostByUser(ctx context.Context, postID, userID string) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", postID, userID).Delete(&models.Post{}).Error
}

func (r PostRepository) GetAllPosts(ctx context.Context) ([]models.Post, error) {
	var posts []models.Post

	if err := r.db.WithContext(ctx).Find(&posts).Error; err != nil {
		return nil, err
	}

	return posts, nil
}
