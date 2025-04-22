package repository

import "gorm.io/gorm"

type UserStatusRepository struct {
	DB *gorm.DB
}
