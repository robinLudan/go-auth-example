package storage

import (
	"errors"

	"github.com/robinLudan/user-auth/internal/models"
)

var (
	ErrEmailExists  = errors.New("email already exists")
	ErrUserNotFound = errors.New("user not found")
)

type Storage interface {
	Register(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	Login(loginReq *models.LoginUserReq) error
}
