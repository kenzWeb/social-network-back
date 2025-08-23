package repository

import "gorm.io/gorm"

type Models struct {
	DB                *gorm.DB
	Users             UserRepository
	VerificationCodes VerificationCodeRepository
	Posts             PostRepository
	Comments          CommentRepository
	Likes             LikeRepository
	Follows           FollowRepository
}

func NewModels(db *gorm.DB) *Models {
	return &Models{
		DB:                db,
		Users:             UserRepository{db: db},
		VerificationCodes: VerificationCodeRepository{db: db},
		Posts:             PostRepository{db: db},
		Comments:          CommentRepository{db: db},
		Likes:             LikeRepository{db: db},
		Follows:           FollowRepository{db: db},
	}
}
