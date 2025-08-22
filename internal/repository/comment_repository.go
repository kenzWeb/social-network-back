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
	var exists bool
	if err := r.db.WithContext(ctx).
		Model(&models.Post{}).
		Select("count(*) > 0").
		Where("id = ?", postId).
		Find(&exists).Error; err != nil {
		return nil, err
	}
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}

	var comments []*models.Comment
	if err := r.db.WithContext(ctx).Where("post_id = ?", postId).Order("created_at ASC").Find(&comments).Error; err != nil {
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

func (r CommentRepository) GetCommentsByUser(ctx context.Context, userId string) ([]*models.Comment, error) {
	var comments []*models.Comment
	if err := r.db.WithContext(ctx).Where("user_id = ?", userId).Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

func (r CommentRepository) CreateComment(ctx context.Context, comment *models.Comment) error {
	// Ensure the post exists before creating a comment
	var exists bool
	if err := r.db.WithContext(ctx).
		Model(&models.Post{}).
		Select("count(*) > 0").
		Where("id = ?", comment.PostID).
		Find(&exists).Error; err != nil {
		return err
	}
	if !exists {
		return gorm.ErrRecordNotFound
	}
	return r.db.WithContext(ctx).Create(comment).Error
}
