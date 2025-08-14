package users

import "gorm.io/gorm"

type User struct {
	ID    string `gorm:"primaryKey"`
	Name  string `gorm:"not null"`
	Email string `gorm:"not null;uniqueIndex"`
}

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&User{})
}
