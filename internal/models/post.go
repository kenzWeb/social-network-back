package models

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User     User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Likes    []Like    `gorm:"foreignKey:PostID" json:"likes,omitempty"`
	Comments []Comment `gorm:"foreignKey:PostID" json:"comments,omitempty"`
}