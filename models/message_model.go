package models

import (
	"time"

	"gorm.io/gorm"
)

type MessageModels struct {
	ID        string    `gorm:"type:uuid;primaryKey;" json:"id"`
	RoomID    string    `gorm:"type:uuid;not null;references:rooms(id)" json:"room_id"`
	SenderID  string    `gorm:"type:uuid;not null;references:users(id)" json:"sender_id"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`
}

// BeforeCreate is a GORM hook that runs before creating a new record

func MigrateMessageModel(db *gorm.DB) error {
	return db.AutoMigrate(&MessageModels{})
}
