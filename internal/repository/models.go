package repository

import "gorm.io/gorm"

// Models агрегирует репозитории.
// Добавляйте сюда новые репозитории/сервисы по мере роста проекта.
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
