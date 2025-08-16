package repository

import (
	"modern-social-media/internal/models"

	"gorm.io/gorm"
)

func MigrateDatabase(db *gorm.DB) error {

	err := db.AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.Story{},
		&models.Like{},
		&models.Comment{},
		&models.Follow{},
	)
	if err != nil {
		return err
	}

	err = createIndexes(db)
	if err != nil {
		return err
	}

	err = createConstraints(db)
	if err != nil {
		return err
	}

	return nil
}

func createIndexes(db *gorm.DB) error {

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_likes_user_id ON likes(user_id)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_likes_post_id ON likes(post_id)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_likes_story_id ON likes(story_id)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_follows_follower_id ON follows(follower_id)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_follows_following_id ON follows(following_id)").Error; err != nil {
		return err
	}

	return nil
}

func createConstraints(db *gorm.DB) error {
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS unique_follows_follower_following ON follows(follower_id, following_id)").Error; err != nil {
		return err
	}

	return nil
}
