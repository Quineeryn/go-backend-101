package users

import "time"

type User struct {
	ID           string    `json:"id" gorm:"type:text;primaryKey"`
	Name         string    `json:"name" gorm:"type:text;not null"`
	Email        string    `json:"email" gorm:"type:text;not null;uniqueIndex"`
	PasswordHash *string   `json:"-" gorm:"column:password_hash"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
