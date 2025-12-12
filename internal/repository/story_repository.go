package repository

import (
	"context"
	"modern-social-media/internal/models"
	"time"

	"gorm.io/gorm"
)

type StoryRepository struct {
	db *gorm.DB
}

func (r StoryRepository) GetAllStories(ctx context.Context, hoursLimit int) ([]models.Story, error) {
	var stories []models.Story
	query := r.db.WithContext(ctx).Preload("User")

	if hoursLimit > 0 {
		timeLimit := time.Now().Add(-time.Duration(hoursLimit) * time.Hour)
		query = query.Where("created_at > ?", timeLimit)
	}

	if err := query.Order("created_at DESC").Find(&stories).Error; err != nil {
		return nil, err
	}
	return stories, nil
}

func (r StoryRepository) GetById(ctx context.Context, id string) (*models.Story, error) {
	var story models.Story
	if err := r.db.WithContext(ctx).
		Preload("User").
		First(&story, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &story, nil
}

func (r StoryRepository) GetStoriesByUser(ctx context.Context, userID string) ([]models.Story, error) {
	var stories []models.Story
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&stories).Error; err != nil {
		return nil, err
	}
	return stories, nil
}

func (r StoryRepository) GetRecentStoriesByUser(ctx context.Context, userID string) ([]models.Story, error) {
	var stories []models.Story
	timeLimit := time.Now().Add(-24 * time.Hour)

	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("user_id = ? AND created_at > ?", userID, timeLimit).
		Order("created_at DESC").
		Find(&stories).Error; err != nil {
		return nil, err
	}
	return stories, nil
}

func (r StoryRepository) GetStoriesFromFollowing(ctx context.Context, userID string) ([]models.Story, error) {
	var stories []models.Story
	timeLimit := time.Now().Add(-24 * time.Hour)

	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("user_id IN (SELECT following_id FROM follows WHERE follower_id = ?) AND created_at > ?", userID, timeLimit).
		Order("created_at DESC").
		Find(&stories).Error; err != nil {
		return nil, err
	}
	return stories, nil
}

func (r StoryRepository) CreateStory(ctx context.Context, story *models.Story) error {
	return r.db.WithContext(ctx).Create(story).Error
}

func (r StoryRepository) UpdateStoryByUser(ctx context.Context, storyID, userID string, story *models.Story) error {
	var existingStory models.Story
	if err := r.db.WithContext(ctx).
		First(&existingStory, "id = ? AND user_id = ?", storyID, userID).Error; err != nil {
		return err
	}

	existingStory.MediaURL = story.MediaURL
	existingStory.MediaType = story.MediaType
	return r.db.WithContext(ctx).Save(&existingStory).Error
}

func (r StoryRepository) DeleteStoryByUser(ctx context.Context, storyID, userID string) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", storyID, userID).
		Delete(&models.Story{})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r StoryRepository) DeleteExpiredStories(ctx context.Context, hoursLimit int) error {
	timeLimit := time.Now().Add(-time.Duration(hoursLimit) * time.Hour)
	return r.db.WithContext(ctx).
		Where("created_at < ?", timeLimit).
		Delete(&models.Story{}).Error
}

func (r StoryRepository) GetFollowedUsersWithStories(ctx context.Context, userID string) ([]models.User, error) {
	var users []models.User
	timeLimit := time.Now().Add(-24 * time.Hour)

	if err := r.db.WithContext(ctx).
		Distinct("users.*").
		Joins("JOIN follows ON follows.following_id = users.id").
		Where("follows.follower_id = ?", userID).
		Where("EXISTS (SELECT 1 FROM stories WHERE stories.user_id = users.id AND stories.created_at > ?)", timeLimit).
		Preload("Stories", "created_at > ?", timeLimit, func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r StoryRepository) GetExpiredStories(ctx context.Context, hoursLimit int) ([]models.Story, error) {
	var stories []models.Story
	timeLimit := time.Now().Add(-time.Duration(hoursLimit) * time.Hour)

	if err := r.db.WithContext(ctx).
		Where("created_at < ?", timeLimit).
		Find(&stories).Error; err != nil {
		return nil, err
	}
	return stories, nil
}
