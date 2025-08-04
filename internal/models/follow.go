package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type Follow struct {
	ID          string `gorm:"type:varchar(25);primaryKey" json:"id"`
	FollowerID  string `gorm:"type:varchar(25);not null;check:follower_id != following_id" json:"follower_id"`
	FollowingID string `gorm:"type:varchar(25);not null" json:"following_id"`
	CreatedAt   time.Time `json:"created_at"`

	Follower  User `gorm:"foreignKey:FollowerID" json:"follower,omitempty"`
	Following User `gorm:"foreignKey:FollowingID" json:"following,omitempty"`
}

// BeforeCreate автоматически генерирует CUID перед созданием
func (f *Follow) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = cuid.New()
	}
	return nil
}

// Уникальный индекс для пары (FollowerID, FollowingID)
func (Follow) TableName() string {
	return "follows"
}
