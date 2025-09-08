package users

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestRepository_Create_And_GetByEmail(t *testing.T) {
	db := newTestDB(t)
	repo := NewRepository(db)

	u := &User{ID: uuid.NewString(), Name: "X", Email: "x@example.com"}
	if err := repo.Create(context.Background(), u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByEmail(context.Background(), "x@example.com")
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if got.Email != "x@example.com" {
		t.Fatalf("want email x@example.com, got %s", got.Email)
	}
}
