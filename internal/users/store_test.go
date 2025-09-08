package users

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- sentinel guard (kalau di package sudah ada, ini tidak dipakai) ---
var (
	_ErrDuplicate = errors.New("duplicate")
	_ErrNotFound  = errors.New("not found")
)

func isErrDuplicate(err error) bool {
	// Pakai yang dari package kalau ada
	if ErrDuplicate != nil && !errors.Is(ErrDuplicate, nil) {
		return errors.Is(err, ErrDuplicate)
	}
	return errors.Is(err, _ErrDuplicate) || strings.Contains(strings.ToLower(err.Error()), "duplicate")
}

func isErrNotFound(err error) bool {
	if ErrNotFound != nil && !errors.Is(ErrNotFound, nil) {
		return errors.Is(err, ErrNotFound)
	}
	return errors.Is(err, _ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound)
}

// --- helpers ---

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// DB in-memory unik per test name (hindari berbagi cache)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB(): %v", err)
	}
	// Pastikan 1 koneksi supaya gak bikin DB memory baru diam-diam
	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Model kamu gak pakai unique tag → bikin index unik manual
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uix_users_email ON users(email);`).Error; err != nil {
		t.Fatalf("create unique index: %v", err)
	}

	return db
}

func newStore(t *testing.T) *Store {
	t.Helper()
	return NewStore(newTestDB(t))
}

// --- tests ---

func TestStore_Create_Success_Normalizes(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	in := User{
		ID:    uuid.NewString(),
		Name:  "  Alea  ",
		Email: "  ALEA@EXAMPLE.COM  ",
	}
	got, err := s.Create(ctx, in)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.ID == "" {
		t.Fatalf("expected ID set, got empty")
	}
	if got.Name != "Alea" {
		t.Fatalf("want Name trimmed, got %q", got.Name)
	}
	if got.Email != "alea@example.com" {
		t.Fatalf("want Email lower+trimmed, got %q", got.Email)
	}
}

func TestStore_Create_Duplicate_ReturnsErrDuplicate(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	if _, err := s.Create(ctx, User{ID: uuid.NewString(), Name: "A", Email: "dup@example.com"}); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	_, err := s.Create(ctx, User{ID: uuid.NewString(), Name: "B", Email: "dup@example.com"})
	if !isErrDuplicate(err) {
		t.Fatalf("want ErrDuplicate, got %v", err)
	}
}

func TestStore_Get_NotFound_ReturnsErrNotFound(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	_, err := s.Get(ctx, "does-not-exist")
	if !isErrNotFound(err) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestStore_List_AscendingByCreatedAt(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	u1, _ := s.Create(ctx, User{ID: uuid.NewString(), Name: "A", Email: "a@example.com"})
	// jeda kecil supaya CreatedAt berbeda
	time.Sleep(2 * time.Millisecond)
	u2, _ := s.Create(ctx, User{ID: uuid.NewString(), Name: "B", Email: "b@example.com"})

	list, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2, got %d", len(list))
	}
	// order ASC by created_at → u1 lalu u2
	if list[0].ID != u1.ID || list[1].ID != u2.ID {
		t.Fatalf("wrong order: %+v", list)
	}
}

func TestStore_Update_Success_AppliesNewValues(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	u, err := s.Create(ctx, User{ID: uuid.NewString(), Name: "Old", Email: "old@example.com"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	upd := User{
		Name:  "  New Name  ",
		Email: "  NEW@EXAMPLE.COM ",
	}
	got, err := s.Update(ctx, u.ID, upd)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got.Name != "New Name" {
		t.Fatalf("want Name updated+trimmed, got %q", got.Name)
	}
	if got.Email != "new@example.com" {
		t.Fatalf("want Email updated lower+trimmed, got %q", got.Email)
	}
}

func TestStore_Update_DuplicateEmail_ReturnsErrDuplicate(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	_, _ = s.Create(ctx, User{ID: uuid.NewString(), Name: "A", Email: "a@example.com"})
	u2, _ := s.Create(ctx, User{ID: uuid.NewString(), Name: "B", Email: "b@example.com"})

	_, err := s.Update(ctx, u2.ID, User{
		Name:  "B2",
		Email: "a@example.com", // nabrak u1
	})
	if !isErrDuplicate(err) {
		t.Fatalf("want ErrDuplicate, got %v", err)
	}
}

func TestStore_Update_EmptyFields_ReturnsError(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	u, _ := s.Create(ctx, User{ID: uuid.NewString(), Name: "X", Email: "x@example.com"})
	_, err := s.Update(ctx, u.ID, User{
		Name:  "   ",
		Email: "   ",
	})
	if err == nil {
		t.Fatalf("want error for empty name/email, got nil")
	}
}

func TestStore_Delete_NotFound(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	err := s.Delete(ctx, "nope")
	if !isErrNotFound(err) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestStore_Delete_Success_Then_Get_NotFound(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	u, _ := s.Create(ctx, User{ID: uuid.NewString(), Name: "Del", Email: "del@example.com"})
	if err := s.Delete(ctx, u.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := s.Get(ctx, u.ID)
	if !isErrNotFound(err) {
		t.Fatalf("want ErrNotFound after delete, got %v", err)
	}
}
