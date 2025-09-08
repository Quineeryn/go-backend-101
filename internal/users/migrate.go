package users

import "gorm.io/gorm"

// AutoMigrate hanya menambah/mengubah skema yang aman (idempotent).
// Jangan melakukan DROP/RENAME/ALTER berisiko di sini.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}
