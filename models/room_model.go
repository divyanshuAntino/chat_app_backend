package models

import (
	"time"

	"gorm.io/gorm"
)

type RoomModels struct {
	ID        string    `gorm:"type:uuid;primaryKey;" json:"id"`
	UserId1   string    `json:"userid1" gorm:"type:uuid;not null"`
	UserId2   string    `json:"userid2" gorm:"type:uuid;not null"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`
}

func MigrateRoom(db *gorm.DB) error {
	return db.AutoMigrate(&RoomModels{})
}
