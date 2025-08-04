package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type Post struct {
	ID            string `gorm:"type:varchar(25);primaryKey" json:"id"`
	UserID        string `gorm:"type:varchar(25);not null" json:"user_id"`
	Content       string `gorm:"type:text;not null" json:"content"`
	ImageURL      string `gorm:"size:255" json:"image_url"`
	LikesCount    int    `gorm:"default:0" json:"likes_count"`
	CommentsCount int    `gorm:"default:0" json:"comments_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	User     User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Likes    []Like    `gorm:"foreignKey:PostID" json:"likes,omitempty"`
	Comments []Comment `gorm:"foreignKey:PostID" json:"comments,omitempty"`
}

// BeforeCreate автоматически генерирует CUID перед созданием
func (p *Post) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = cuid.New()
	}
	return nil
}