package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type Like struct {
	ID      string  `gorm:"type:varchar(25);primaryKey" json:"id"`
	UserID  string  `gorm:"type:varchar(25);not null;index:user_post_like,unique;index:user_story_like,unique" json:"user_id"`
	PostID  *string `gorm:"type:varchar(25);index:user_post_like,unique" json:"post_id,omitempty"`
	StoryID *string `gorm:"type:varchar(25);index:user_story_like,unique" json:"story_id,omitempty"`

	CreatedAt time.Time `json:"created_at"`

	User  User   `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Post  *Post  `gorm:"foreignKey:PostID;references:ID" json:"post,omitempty"`
	Story *Story `gorm:"foreignKey:StoryID;references:ID" json:"story,omitempty"`
}

func (l *Like) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = cuid.New()
	}
	return nil
}

func (Like) TableName() string {
	return "likes"
}
