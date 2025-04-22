package models

import (
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := MigrateUser(db); err != nil {
		return err
	}
	if err := MigrateRoom(db); err != nil {
		return err
	}
	if err := MigrateMessageModel(db); err != nil {
		return err
	}
	if err := MigrateUserStatus(db); err != nil {
		return err
	}
	return nil
}
