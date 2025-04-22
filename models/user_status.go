package models

import (
	"time"

	"gorm.io/gorm"
)

type UserStatusModles struct {
	UserID    string    `gorm:"type:uuid;primaryKey;references:users(id)" json:"user_id"`
	IsOnline  bool      `gorm:"default:false" json:"is_online"`
	LastSeen  time.Time `gorm:"default:now()" json:"last_seen"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`
}

func MigrateUserStatus(db *gorm.DB) error {
	return db.AutoMigrate(&UserStatusModles{})
}
