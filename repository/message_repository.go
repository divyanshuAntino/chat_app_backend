package repository

import "gorm.io/gorm"

type MessageRepository struct {
	DB *gorm.DB
}
