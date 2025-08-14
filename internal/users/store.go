package users

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrNotFound  = errors.New("user not found")
	ErrDuplicate = errors.New("duplicate key")
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(u User) (User, error) {
	if err := s.db.Create(&u).Error; err != nil {
		// gorm akan memetakan unique violation ke gorm.ErrDuplicatedKey (pg/sqlite)
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return User{}, ErrDuplicate
		}
		return User{}, err
	}
	return u, nil
}

func (s *Store) List() []User {
	var users []User
	s.db.Find(&users)
	return users
}

func (s *Store) Get(id string) (User, error) {
	var u User
	if err := s.db.First(&u, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	return u, nil
}

func (s *Store) Update(id string, data User) (User, error) {
	var u User
	if err := s.db.First(&u, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	u.Name = data.Name
	u.Email = data.Email
	if err := s.db.Save(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return User{}, ErrDuplicate
		}
		return User{}, err
	}
	return u, nil
}

func (s *Store) Delete(id string) error {
	if err := s.db.Delete(&User{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}
