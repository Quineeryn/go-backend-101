package users

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgconn"
	"gorm.io/gorm"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

// isDuplicateErr tries to normalize unique-violation across drivers.
func isDuplicateErr(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}

	// 2) Postgres (pgx/pgconn) → kode 23505 = unique_violation
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return true
		}
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "sqlstate 23505") ||
		strings.Contains(msg, "duplicate key value violates unique constraint") ||
		strings.Contains(msg, "unique constraint failed")
}

func (s *Store) Create(ctx context.Context, u User) (User, error) {
	u.Name = strings.TrimSpace(u.Name)
	u.Email = strings.TrimSpace(u.Email)
	u.Name = strings.TrimSpace(u.Name)
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))

	if err := s.db.WithContext(ctx).Create(&u).Error; err != nil {
		if isDuplicateErr(err) {
			return User{}, ErrDuplicate
		}
		return User{}, err
	}
	return u, nil
}

func (s *Store) List(ctx context.Context) ([]User, error) {
	var users []User
	if err := s.db.WithContext(ctx).
		Select("id", "name", "email", "created_at").
		Order("created_at ASC").
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Store) Get(ctx context.Context, id string) (User, error) {
	var u User
	if err := s.db.WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	return u, nil
}

func (s *Store) Update(ctx context.Context, id string, data User) (User, error) {
	var u User
	if err := s.db.WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}

	data.Name = strings.TrimSpace(data.Name)
	data.Email = strings.TrimSpace(data.Email)
	data.Name = strings.TrimSpace(u.Name)
	data.Email = strings.ToLower(strings.TrimSpace(u.Email))

	if data.Name == "" || data.Email == "" {
		return User{}, errors.New("name and email are required")
	}

	u.Name = data.Name
	u.Email = data.Email

	if err := s.db.WithContext(ctx).Save(&u).Error; err != nil {
		if isDuplicateErr(err) {
			return User{}, ErrDuplicate
		}
		return User{}, err
	}
	return u, nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	res := s.db.WithContext(ctx).Delete(&User{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := s.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}

func (s *Store) FindByID(ctx context.Context, id string) (User, error) {
	var u User
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	return u, err
}
