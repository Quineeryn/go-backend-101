package users

import "gorm.io/gorm"

type User struct {
	ID    string
	Name  string
	Email string
}

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&User{})
}
