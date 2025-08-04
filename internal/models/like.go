package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type Like struct {
	ID        string `gorm:"type:varchar(25);primaryKey" json:"id"`
	UserID    string `gorm:"type:varchar(25);not null" json:"user_id"`
	PostID    string `gorm:"type:varchar(25);not null" json:"post_id"`
	CreatedAt time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Post Post `gorm:"foreignKey:PostID" json:"post,omitempty"`
}

// BeforeCreate автоматически генерирует CUID перед созданием
func (l *Like) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = cuid.New()
	}
	return nil
}

// Уникальный индекс для пары (UserID, PostID)
func (Like) TableName() string {
	return "likes"
}