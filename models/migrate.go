package models

import (
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := MigrateUser(db); err != nil {
		return err
	}

	return nil
}
