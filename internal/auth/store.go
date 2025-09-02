package auth

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type RefreshToken struct {
	ID        string `gorm:"primaryKey"` // uuid
	UserID    string
	JTI       string `gorm:"index"`
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type Store struct{ db *gorm.DB }

func NewStore(db *gorm.DB) *Store { return &Store{db} }

func (s *Store) Save(ctx context.Context, rt *RefreshToken) error {
	return s.db.WithContext(ctx).Create(rt).Error
}

func (s *Store) RevokeByJTI(ctx context.Context, jti string) error {
	now := time.Now().UTC()
	return s.db.WithContext(ctx).
		Model(&RefreshToken{}).
		Where("jti = ? AND revoked_at IS NULL", jti).
		Update("revoked_at", now).Error
}

func (s *Store) IsActive(ctx context.Context, jti string) (bool, error) {
	var rt RefreshToken
	err := s.db.WithContext(ctx).
		Where("jti = ? AND revoked_at IS NULL AND expires_at > ?", jti, time.Now().UTC()).
		First(&rt).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return err == nil, err
}
