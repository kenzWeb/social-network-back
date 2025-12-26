package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type Notification struct {
	ID        string `gorm:"type:varchar(25);primaryKey" json:"id"`
	UserID    string `gorm:"type:varchar(25);not null" json:"user_id"`
	Message   string `gorm:"type:text;not null" json:"message"`
	Read      bool   `gorm:"type:boolean;default:false" json:"read"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = cuid.New()
	}
	return nil
}

func (Notification) TableName() string {
	return "notifications"
}