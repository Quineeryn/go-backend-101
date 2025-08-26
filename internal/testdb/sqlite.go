package testdb

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Open returns a brand-new in-memory SQLite DB for each test.
func Open(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite test db: %v", err)
	}
	return db
}
