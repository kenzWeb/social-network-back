package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type VerificationCode struct {
	ID         string     `gorm:"type:varchar(25);primaryKey" json:"id"`
	UserID     string     `gorm:"type:varchar(25);index;not null" json:"user_id"`
	Purpose    string     `gorm:"size:50;index;not null" json:"purpose"`
	Code       string     `gorm:"size:10;not null" json:"code"`
	ExpiresAt  time.Time  `gorm:"index" json:"expires_at"`
	ConsumedAt *time.Time `json:"consumed_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (v *VerificationCode) BeforeCreate(tx *gorm.DB) error {
	if v.ID == "" {
		v.ID = cuid.New()
	}
	return nil
}
