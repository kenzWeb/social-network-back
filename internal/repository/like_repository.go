package repository

import (
	"modern-social-media/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LikeRepository struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) *LikeRepository {
	return &LikeRepository{db: db}
}


func (r *LikeRepository) CreateLike(like *models.Like) error {
	return r.db.Create(like).Error
}


func (r *LikeRepository) DeleteLike(userID, postID uuid.UUID) error {
	return r.db.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&models.Like{}).Error
}


func (r *LikeRepository) GetLike(userID, postID uuid.UUID) (*models.Like, error) {
	var like models.Like
	err := r.db.Where("user_id = ? AND post_id = ?", userID, postID).First(&like).Error
	if err != nil {
		return nil, err
	}
	return &like, nil
}


func (r *LikeRepository) IsLiked(userID, postID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Like{}).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Count(&count).Error
	return count > 0, err
}


func (r *LikeRepository) GetLikesByPostID(postID uuid.UUID) ([]models.Like, error) {
	var likes []models.Like
	err := r.db.Preload("User").Where("post_id = ?", postID).Find(&likes).Error
	return likes, err
}


func (r *LikeRepository) GetLikesByUserID(userID uuid.UUID, limit, offset int) ([]models.Like, error) {
	var likes []models.Like
	err := r.db.Preload("Post").
		Preload("Post.User").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&likes).Error
	return likes, err
}


func (r *LikeRepository) GetLikesCount(postID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Like{}).Where("post_id = ?", postID).Count(&count).Error
	return count, err
}
