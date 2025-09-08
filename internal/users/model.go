package users

import "time"

type User struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name"`
	Email        string    `json:"email" gorm:"column:email"`
	Role         string    `json:"role" gorm:"default:user"`
	PasswordHash *string   `json:"-" gorm:"column:password_hash"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
