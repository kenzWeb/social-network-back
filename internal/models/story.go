package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type Story struct {
	ID         string    `gorm:"type:varchar(25);primaryKey" json:"id"`
	UserID     string    `gorm:"type:varchar(25);not null" json:"user_id"`
	MediaURL   string    `gorm:"size:255;not null" json:"media_url"`
	MediaType  string    `gorm:"size:20;not null;default:'image'" json:"media_type"`
	LikesCount int       `gorm:"default:0" json:"likes_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	User  User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Likes []Like `gorm:"foreignKey:StoryID;constraint:OnDelete:CASCADE" json:"likes,omitempty"`
}

func (u *Story) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = cuid.New()
	}
	return nil
}
