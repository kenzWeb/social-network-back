package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type Comment struct {
	ID        string `gorm:"type:varchar(25);primaryKey" json:"id"`
	UserID    string `gorm:"type:varchar(25);not null" json:"user_id"`
	PostID    string `gorm:"type:varchar(25);not null" json:"post_id"`
	Message   string `gorm:"type:text;not null" json:"message"`
	ImageURL  string `gorm:"size:255" json:"image_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Post Post `gorm:"foreignKey:PostID" json:"post,omitempty"`
}

// BeforeCreate автоматически генерирует CUID перед созданием
func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = cuid.New()
	}
	return nil
}