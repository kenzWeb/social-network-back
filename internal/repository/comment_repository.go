package repository

import (
	"context"
	"modern-social-media/internal/models"

	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func (r CommentRepository) GetCommentsByPost(ctx context.Context, postId string) ([]*models.Comment, error) {
	var comments []*models.Comment
	if err := r.db.WithContext(ctx).Where("post_id = ?", postId).Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

func (r CommentRepository) GetCommentById(ctx context.Context, id string) (*models.Comment, error) {
	var comment models.Comment
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&comment).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r CommentRepository) CreateComment(ctx context.Context, comment *models.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}
