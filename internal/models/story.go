package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type Story struct {
	ID         string    `gorm:"type:varchar(25);primaryKey" json:"id"`
	UserID     string    `gorm:"type:varchar(25);not null" json:"user_id"`
	ContentURL string    `gorm:"size:255" json:"content_url"`
	LikesCount int       `gorm:"default:0" json:"likes_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Likes []Like `gorm:"foreignKey:StoryID;constraint:OnDelete:CASCADE" json:"likes,omitempty"`
}

func (u *Story) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = cuid.New()
	}
	return nil
}
