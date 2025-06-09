package services

import (
	"errors"
	"auth.alexmust/internal/models"
)

type UserService interface {
	Register(email, password string) error
	Login(email, password string) (*models.User, error)
}

type userService struct {
	users map[string]string // map[email]hashedPassword
}

func NewUserService() UserService {
	return &userService{users: make(map[string]string)}
}

func (u *userService) Register(email, password string) error {
	if _, exists := u.users[email]; exists {
		return errors.New("user already exists")
	}
	u.users[email] = password // хэшируй тут на практике
	return nil
}

func (u *userService) Login(email, password string) (*models.User, error) {
	stored, exists := u.users[email]
	if !exists || stored != password {
		return nil, errors.New("invalid credentials")
	}
	return &models.User{Email: email}, nil
}
