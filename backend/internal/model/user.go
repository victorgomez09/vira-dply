package model

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Username string `gorm:"uniqueIndex;not null"`
	Email    string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
}

func (u *User) TableName() string {
	return "users"
}
