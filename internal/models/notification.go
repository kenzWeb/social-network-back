package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type NotificationType string

const (
	NotificationTypeFollow  NotificationType = "follow"
	NotificationTypeLike    NotificationType = "like"
	NotificationTypeComment NotificationType = "comment"
	NotificationTypeMention NotificationType = "mention"
)

type Notification struct {
	ID        string           `gorm:"type:varchar(25);primaryKey" json:"id"`
	UserID    string           `gorm:"type:varchar(25);not null;index" json:"user_id"`
	ActorID   string           `gorm:"type:varchar(25);not null" json:"actor_id"`
	Type      NotificationType `gorm:"type:varchar(20);not null;index" json:"type"`
	TargetID  *string          `gorm:"type:varchar(25)" json:"target_id,omitempty"`
	Read      bool             `gorm:"type:boolean;default:false;index" json:"read"`
	CreatedAt time.Time        `gorm:"index" json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`

	User  User `gorm:"foreignKey:UserID" json:"-"`
	Actor User `gorm:"foreignKey:ActorID" json:"actor,omitempty"`
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
