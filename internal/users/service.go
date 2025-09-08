package users

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(ctx context.Context, name, email string) (*User, error) {
	user := &User{
		ID:    uuid.New().String(),
		Name:  name,
		Email: email,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// (opsional, buat CP‑05 nanti)
// func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
// 	return s.repo.GetByEmail(ctx, email)
// }
