package models

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	PostID    uuid.UUID `gorm:"type:uuid;not null" json:"post_id"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	ImageURL	string    `gorm:"size:255" json:"image_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Post Post `gorm:"foreignKey:PostID" json:"post,omitempty"`
}