package repository

import (
	"context"
	"errors"
	"modern-social-media/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LikeRepository struct {
	db *gorm.DB
}

func (r LikeRepository) TogglePostLike(ctx context.Context, userID, postID string) (bool, error) {
	return r.toggle(ctx, userID, &postID, nil)
}

func (r LikeRepository) ToggleStoryLike(ctx context.Context, userID, storyID string) (bool, error) {
	return r.toggle(ctx, userID, nil, &storyID)
}

func (r LikeRepository) HasUserLikedPost(ctx context.Context, userID, postID string) (bool, error) {
	var cnt int64
	if err := r.db.WithContext(ctx).Model(&models.Like{}).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func (r LikeRepository) HasUserLikedStory(ctx context.Context, userID, storyID string) (bool, error) {
	var cnt int64
	if err := r.db.WithContext(ctx).Model(&models.Like{}).
		Where("user_id = ? AND story_id = ?", userID, storyID).
		Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func (r LikeRepository) CountPostLikes(ctx context.Context, postID string) (int64, error) {
	var cnt int64
	if err := r.db.WithContext(ctx).Model(&models.Like{}).Where("post_id = ?", postID).Count(&cnt).Error; err != nil {
		return 0, err
	}
	return cnt, nil
}

func (r LikeRepository) CountStoryLikes(ctx context.Context, storyID string) (int64, error) {
	var cnt int64
	if err := r.db.WithContext(ctx).Model(&models.Like{}).Where("story_id = ?", storyID).Count(&cnt).Error; err != nil {
		return 0, err
	}
	return cnt, nil
}

func (r LikeRepository) toggle(ctx context.Context, userID string, postID, storyID *string) (bool, error) {
	var liked bool
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exists bool
		if postID != nil {
			if err := tx.Model(&models.Post{}).Select("count(*) > 0").Where("id = ?", *postID).Find(&exists).Error; err != nil {
				return err
			}
			if !exists {
				return gorm.ErrRecordNotFound
			}
		} else if storyID != nil {
			if err := tx.Model(&models.Story{}).Select("count(*) > 0").Where("id = ?", *storyID).Find(&exists).Error; err != nil {
				return err
			}
			if !exists {
				return gorm.ErrRecordNotFound
			}
		} else {
			return errors.New("either postID or storyID required")
		}

		var l models.Like
		q := tx.Model(&models.Like{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID)
		if postID != nil {
			q = q.Where("post_id = ?", *postID)
		} else {
			q = q.Where("story_id = ?", *storyID)
		}

		if err := q.First(&l).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				like := &models.Like{UserID: userID, PostID: postID, StoryID: storyID}
				if err := tx.Create(like).Error; err != nil {
					return err
				}
				if postID != nil {
					if err := tx.Model(&models.Post{}).Where("id = ?", *postID).UpdateColumn("likes_count", gorm.Expr("likes_count + 1")).Error; err != nil {
						return err
					}
				} else {
					if err := tx.Model(&models.Story{}).Where("id = ?", *storyID).UpdateColumn("likes_count", gorm.Expr("likes_count + 1")).Error; err != nil {
						return err
					}
				}
				liked = true
				return nil
			}
			return err
		}

		if err := tx.Delete(&l).Error; err != nil {
			return err
		}
		if postID != nil {
			if err := tx.Model(&models.Post{}).Where("id = ?", *postID).UpdateColumn("likes_count", gorm.Expr("GREATEST(likes_count - 1, 0)")).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&models.Story{}).Where("id = ?", *storyID).UpdateColumn("likes_count", gorm.Expr("GREATEST(likes_count - 1, 0)")).Error; err != nil {
				return err
			}
		}
		liked = false
		return nil
	})
	return liked, err
}
