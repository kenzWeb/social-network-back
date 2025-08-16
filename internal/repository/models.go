package repository

import "gorm.io/gorm"

type Models struct {
	DB                *gorm.DB
	Users             UserRepository
	VerificationCodes VerificationCodeRepository
}

func NewModels(db *gorm.DB) *Models {
	return &Models{
		DB:                db,
		Users:             UserRepository{db: db},
		VerificationCodes: VerificationCodeRepository{db: db},
	}
}
