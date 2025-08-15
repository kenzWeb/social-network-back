package repository

import "gorm.io/gorm"

type Models struct {
	DB    *gorm.DB
	Users UserRepository
}

func NewModels(db *gorm.DB) *Models {
	return &Models{
		DB:    db,
		Users: UserRepository{db: db},
	}
}
