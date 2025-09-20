package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

type Conversation struct {
	ID        string    `gorm:"type:varchar(25);primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Participants []ConversationParticipant `gorm:"foreignKey:ConversationID" json:"participants,omitempty"`
	Messages     []Message                 `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

func (c *Conversation) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = cuid.New()
	}
	return nil
}

type ConversationParticipant struct {
	ID             string    `gorm:"type:varchar(25);primaryKey" json:"id"`
	ConversationID string    `gorm:"type:varchar(25);index;not null" json:"conversation_id"`
	UserID         string    `gorm:"type:varchar(25);index;not null" json:"user_id"`
	JoinedAt       time.Time `gorm:"autoCreateTime" json:"joined_at"`
	LastReadAt *time.Time `json:"last_read_at"`
}

func (p *ConversationParticipant) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = cuid.New()
	}
	return nil
}

type Message struct {
	ID             string    `gorm:"type:varchar(25);primaryKey" json:"id"`
	ConversationID string    `gorm:"type:varchar(25);index;not null" json:"conversation_id"`
	SenderID       string    `gorm:"type:varchar(25);index;not null" json:"sender_id"`
	Body           string    `gorm:"type:text;not null" json:"body"`
	CreatedAt      time.Time `json:"created_at"`
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = cuid.New()
	}
	return nil
}


// Инструкция:

// 1.Скопировать ссылку, которую я скинул выше.
// 2.В приложении справа сверху нажать на «+» и вставить с буфера обмена 