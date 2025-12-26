package repository

import (
	"modern-social-media/internal/models"

	"gorm.io/gorm"
)

func MigrateDatabase(db *gorm.DB) error {
	err := customMigrations(db)
	if err != nil {
		return err
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.Story{},
		&models.Like{},
		&models.Comment{},
		&models.Follow{},
		&models.VerificationCode{},
		&models.Conversation{},
		&models.ConversationParticipant{},
		&models.Message{},
		&models.Skill{},
		&models.Notification{},
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

	// Chat indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_conv_participants_user ON conversation_participants(user_id)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at DESC)").Error; err != nil {
		return err
	}

	// Notification indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, read)").Error; err != nil {
		return err
	}

	return nil
}

func customMigrations(db *gorm.DB) error {
	var tableExists bool
	err := db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'stories')").Scan(&tableExists).Error
	if err != nil {
		return err
	}

	if tableExists {
		var columnExists bool
		err = db.Raw("SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'stories' AND column_name = 'image_url')").Scan(&columnExists).Error
		if err != nil {
			return err
		}

		if columnExists {
			if err := db.Exec("ALTER TABLE stories ADD COLUMN IF NOT EXISTS media_url varchar(255)").Error; err != nil {
				return err
			}
			if err := db.Exec("ALTER TABLE stories ADD COLUMN IF NOT EXISTS media_type varchar(20) DEFAULT 'image'").Error; err != nil {
				return err
			}

			if err := db.Exec("UPDATE stories SET media_url = image_url WHERE media_url IS NULL AND image_url IS NOT NULL").Error; err != nil {
				return err
			}

			if err := db.Exec("UPDATE stories SET media_type = 'image' WHERE media_type IS NULL").Error; err != nil {
				return err
			}

			if err := db.Exec("ALTER TABLE stories DROP COLUMN IF EXISTS image_url").Error; err != nil {
				return err
			}

			if err := db.Exec("ALTER TABLE stories ALTER COLUMN media_url SET NOT NULL").Error; err != nil {
				return err
			}
			if err := db.Exec("ALTER TABLE stories ALTER COLUMN media_type SET NOT NULL").Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func createConstraints(db *gorm.DB) error {
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS unique_follows_follower_following ON follows(follower_id, following_id)").Error; err != nil {
		return err
	}

	return nil
}
