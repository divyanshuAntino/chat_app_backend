package models

import "gorm.io/gorm"

type UserModels struct {
	UserId       string  `json:"userid" gorm:"type:uuid;primaryKey"`
	UserEmail    *string `json:"useremail" gorm:"type:varchar(255);not null;unique"`
	UserName     *string `json:"username" gorm:"type:varchar(255);"`
	UserPassword *string `json:"userpassword" gorm:"type:varchar(255);not null"`
	UserImage    *string `json:"userimage" gorm:"type:varchar(255);"`
	Name         *string `json:"name" gorm:"type:varchar(255)"`
	Dob          *string `json:"dob" gorm:"type:date;"`
	TagLine      *string `json:"tagline" gorm:"type:varchar(255);"`
}

func MigrateUser(db *gorm.DB) error {
	return db.AutoMigrate(&UserModels{})
}
