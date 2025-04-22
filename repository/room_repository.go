package repository

import "gorm.io/gorm"

type RoomRepository struct {
	DB *gorm.DB
}
