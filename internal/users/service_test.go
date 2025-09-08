package users

import (
	"context"
	"errors"
	"testing"
)

type mockRepo struct {
	createFn func(ctx context.Context, u *User) error
}

func (m *mockRepo) Create(ctx context.Context, u *User) error {
	return m.createFn(ctx, u)
}
func (m *mockRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	return nil, ErrNotFound
}

func TestService_CreateUser_Success(t *testing.T) {
	repo := &mockRepo{
		createFn: func(ctx context.Context, u *User) error {
			return nil
		},
	}
	svc := NewService(repo)

	u, err := svc.CreateUser(context.Background(), "Alea", "alea@example.com")
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if u.ID == "" {
		t.Fatalf("want ID generated, got empty")
	}
}

func TestService_CreateUser_Duplicate(t *testing.T) {
	repo := &mockRepo{
		createFn: func(ctx context.Context, u *User) error {
			return ErrDuplicate
		},
	}
	svc := NewService(repo)

	_, err := svc.CreateUser(context.Background(), "Alea", "alea@example.com")
	if !errors.Is(err, ErrDuplicate) {
		t.Fatalf("want ErrDuplicate, got %v", err)
	}
}
