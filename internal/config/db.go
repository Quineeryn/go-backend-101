package config

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func OpenDB(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("cannot connect to db: %v", err)
	}
	return db
}
