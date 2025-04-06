package models

import "gorm.io/gorm"

type UserModels struct {
	UserId       string  `json:"userid" grom:"type:uuid;primary_key"`
	UserEmail    *string `json:"useremail" gorm:"type:varchar(255);not null;unique"`
	UserName     *string `json:"username" gorm:"type:varchar(255);not null"`
	UserPassword *string `json:"userpassword" gorm:"type:varchar(255);not null"`
}

func MigrateUser(db *gorm.DB) error {
	return db.AutoMigrate(&UserModels{})
}
