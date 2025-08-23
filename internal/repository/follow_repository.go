package repository

import (
	"context"
	"errors"
	"modern-social-media/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FollowRepository struct {
	db *gorm.DB
}

func (r FollowRepository) ToggleFollow(ctx context.Context, followerID, followingID string) (bool, error) {
	if followerID == followingID {
		return false, errors.New("cannot follow yourself")
	}
	var nowFollowing bool
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exists int64
		if err := tx.Model(&models.User{}).Where("id = ?", followingID).Count(&exists).Error; err != nil {
			return err
		}
		if exists == 0 {
			return gorm.ErrRecordNotFound
		}

		var rel models.Follow
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("follower_id = ? AND following_id = ?", followerID, followingID).First(&rel).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				f := &models.Follow{FollowerID: followerID, FollowingID: followingID}
				res := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "follower_id"}, {Name: "following_id"}},
					DoNothing: true,
				}).Create(f)
				if res.Error != nil {
					return res.Error
				}
				nowFollowing = true
				return nil
			}
			return err
		}
		if err := tx.Delete(&rel).Error; err != nil {
			return err
		}
		nowFollowing = false
		return nil
	})
	return nowFollowing, err
}

func (r FollowRepository) Follow(ctx context.Context, followerID, followingID string) (bool, error) {
	if followerID == followingID {
		return false, errors.New("cannot follow yourself")
	}
	f := models.Follow{FollowerID: followerID, FollowingID: followingID}
	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "follower_id"}, {Name: "following_id"}},
		DoNothing: true,
	}).Create(&f)
	return res.RowsAffected == 1, res.Error
}

func (r FollowRepository) Unfollow(ctx context.Context, followerID, followingID string) (bool, error) {
	res := r.db.WithContext(ctx).Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&models.Follow{})
	return res.RowsAffected == 1, res.Error
}

func (r FollowRepository) IsFollowing(ctx context.Context, followerID, followingID string) (bool, error) {
	var cnt int64
	if err := r.db.WithContext(ctx).Model(&models.Follow{}).Where("follower_id = ? AND following_id = ?", followerID, followingID).Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func (r FollowRepository) GetFollowers(ctx context.Context, userID string, limit, offset int) ([]models.User, error) {
	var users []models.User
	q := r.db.WithContext(ctx).Model(&models.User{}).
		Joins("JOIN follows f ON f.follower_id = users.id").
		Where("f.following_id = ?", userID).
		Order("f.created_at DESC").
		Limit(limit).Offset(offset)
	if err := q.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r FollowRepository) GetFollowing(ctx context.Context, userID string, limit, offset int) ([]models.User, error) {
	var users []models.User
	q := r.db.WithContext(ctx).Model(&models.User{}).
		Joins("JOIN follows f ON f.following_id = users.id").
		Where("f.follower_id = ?", userID).
		Order("f.created_at DESC").
		Limit(limit).Offset(offset)
	if err := q.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r FollowRepository) CountFollowers(ctx context.Context, userID string) (int64, error) {
	var cnt int64
	if err := r.db.WithContext(ctx).Model(&models.Follow{}).Where("following_id = ?", userID).Count(&cnt).Error; err != nil {
		return 0, err
	}
	return cnt, nil
}

func (r FollowRepository) CountFollowing(ctx context.Context, userID string) (int64, error) {
	var cnt int64
	if err := r.db.WithContext(ctx).Model(&models.Follow{}).Where("follower_id = ?", userID).Count(&cnt).Error; err != nil {
		return 0, err
	}
	return cnt, nil
}
