package storage

import (
	"errors"

	"github.com/robinLudan/user-auth/internal/models"
)

var ErrEmailExists = errors.New("User not found")

type Storage interface {
	Register(user *models.User) error
}
