package users

import (
	"errors"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("user not found")

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(user User) User {
	s.db.Create(&user)
	return user
}

func (s *Store) Get(id string) (User, error) {
	var user User
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return user, ErrNotFound
		}
		return user, err
	}
	return user, nil
}

func (s *Store) List() []User {
	var users []User
	s.db.Find(&users)
	return users
}

func (s *Store) Update(id string, data User) (User, error) {
	var u User
	if err := s.db.First(&u, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return u, ErrNotFound
		}
		return u, err
	}
	u.Name = data.Name
	u.Email = data.Email
	s.db.Save(&u)
	return u, nil
}

func (s *Store) Delete(id string) error {
	if err := s.db.Delete(&User{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}
