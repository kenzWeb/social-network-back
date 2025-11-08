package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type User struct {
	ID           string    `gorm:"type:varchar(25);primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Email        string    `gorm:"uniqueIndex;not null;size:100" json:"email"`
	Password     string    `gorm:"not null;size:255" json:"-"`
	FirstName    string    `gorm:"size:50" json:"first_name"`
	LastName     string    `gorm:"size:50" json:"last_name"`
	Bio          string    `gorm:"type:text" json:"bio"`
	AvatarURL    string    `gorm:"size:255" json:"avatar_url"`
	IsVerified   bool      `gorm:"default:false" json:"is_verified"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	Is2FAEnabled bool      `gorm:"default:false" json:"is_2fa_enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	FollowersCount int64 `gorm:"-" json:"followers_count,omitempty"`
	FollowingCount int64 `gorm:"-" json:"following_count,omitempty"`

	Posts     []Post    `gorm:"foreignKey:UserID" json:"posts,omitempty"`
	Likes     []Like    `gorm:"foreignKey:UserID" json:"likes,omitempty"`
	Comments  []Comment `gorm:"foreignKey:UserID" json:"comments,omitempty"`
	Followers []Follow  `gorm:"foreignKey:FollowingID" json:"followers,omitempty"`
	Following []Follow  `gorm:"foreignKey:FollowerID" json:"following,omitempty"`
	Stories   []Story   `gorm:"foreignKey:UserID" json:"stories"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = cuid.New()
	}
	return nil
}
