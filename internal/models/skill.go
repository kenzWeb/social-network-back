package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type Skill struct {
	ID        string    `gorm:"type:varchar(25);primaryKey" json:"id"`
	UserID    string    `gorm:"type:varchar(25);index;not null" json:"user_id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (u *Skill) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = cuid.New()
	}
	return nil
}
