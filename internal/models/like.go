package models

import (
	"time"

	"github.com/google/uuid"
)

type Like struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	PostID    uuid.UUID `gorm:"type:uuid;not null" json:"post_id"`
	CreatedAt time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Post Post `gorm:"foreignKey:PostID" json:"post,omitempty"`
}