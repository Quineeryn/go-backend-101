package users

import (
	"errors"
	"testing"

	"github.com/jackc/pgconn"
	"gorm.io/gorm"
)

func TestIsDuplicateErr_GormDuplicatedKey(t *testing.T) {
	// Simulasi gorm.ErrDuplicatedKey
	if !isDuplicateErr(gorm.ErrDuplicatedKey) {
		t.Fatalf("want true for gorm.ErrDuplicatedKey")
	}
}

func TestIsDuplicateErr_PgError23505(t *testing.T) {
	// Simulasi pg error 23505
	pgErr := &pgconn.PgError{Code: "23505"}
	if !isDuplicateErr(pgErr) {
		t.Fatalf("want true for pg 23505")
	}
}

func TestIsDuplicateErr_ByMessage(t *testing.T) {
	// Simulasi driver lain: deteksi via pesan
	err := errors.New("duplicate key value violates unique constraint \"users_email_key\"")
	if !isDuplicateErr(err) {
		t.Fatalf("want true for duplicate message")
	}
}
