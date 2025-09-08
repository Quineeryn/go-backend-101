package users

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestStore_FindByID_Found(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	u, err := s.Create(ctx, User{ID: uuid.NewString(), Name: "A", Email: "a@example.com"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.FindByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID != u.ID || got.Email != "a@example.com" {
		t.Fatalf("mismatch: %+v", got)
	}
}

func TestStore_FindByID_NotFound(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	if _, err := s.FindByID(ctx, "nope"); err == nil {
		t.Fatalf("want error for not found, got nil")
	}
}
