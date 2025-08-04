package repository

import (
	"modern-social-media/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}


func (r *PostRepository) CreatePost(post *models.Post) error {
	return r.db.Create(post).Error
}


func (r *PostRepository) GetPostByID(id uuid.UUID) (*models.Post, error) {
	var post models.Post
	err := r.db.Preload("User").Where("id = ?", id).First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}


func (r *PostRepository) GetPostsByUserID(userID uuid.UUID, limit, offset int) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.Preload("User").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&posts).Error
	return posts, err
}


func (r *PostRepository) UpdatePost(post *models.Post) error {
	return r.db.Save(post).Error
}


func (r *PostRepository) DeletePost(id uuid.UUID) error {
	return r.db.Delete(&models.Post{}, id).Error
}


func (r *PostRepository) GetAllPosts(limit, offset int) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.Preload("User").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&posts).Error
	return posts, err
}


func (r *PostRepository) GetPostWithComments(id uuid.UUID) (*models.Post, error) {
	var post models.Post
	err := r.db.Preload("User").
		Preload("Comments").
		Preload("Comments.User").
		Where("id = ?", id).First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}


func (r *PostRepository) GetPostWithLikes(id uuid.UUID) (*models.Post, error) {
	var post models.Post
	err := r.db.Preload("User").
		Preload("Likes").
		Preload("Likes.User").
		Where("id = ?", id).First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}


func (r *PostRepository) SearchPosts(query string, limit, offset int) ([]models.Post, error) {
	var posts []models.Post
	searchQuery := "%" + query + "%"
	err := r.db.Preload("User").
		Where("content ILIKE ?", searchQuery).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&posts).Error
	return posts, err
}


func (r *PostRepository) IncrementLikesCount(postID uuid.UUID) error {
	return r.db.Model(&models.Post{}).
		Where("id = ?", postID).
		Update("likes_count", gorm.Expr("likes_count + 1")).Error
}


func (r *PostRepository) DecrementLikesCount(postID uuid.UUID) error {
	return r.db.Model(&models.Post{}).
		Where("id = ?", postID).
		Update("likes_count", gorm.Expr("likes_count - 1")).Error
}


func (r *PostRepository) IncrementCommentsCount(postID uuid.UUID) error {
	return r.db.Model(&models.Post{}).
		Where("id = ?", postID).
		Update("comments_count", gorm.Expr("comments_count + 1")).Error
}


func (r *PostRepository) DecrementCommentsCount(postID uuid.UUID) error {
	return r.db.Model(&models.Post{}).
		Where("id = ?", postID).
		Update("comments_count", gorm.Expr("comments_count - 1")).Error
}
