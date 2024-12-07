package storage

import (
	"errors"

	"github.com/google/uuid"
	"github.com/robinLudan/user-auth/internal/models"
)

var (
	ErrEmailExists  = errors.New("email already exists")
	ErrUserNotFound = errors.New("user not found")
)

type Storage interface {
	Register(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	Login(loginReq *models.LoginUserReq) error
}
