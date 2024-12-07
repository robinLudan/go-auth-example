package storage

import (
	"database/sql"
	"log"

	"github.com/google/uuid"
	"github.com/robinLudan/user-auth/internal/models"
)

type SQLite struct {
	db *sql.DB
}

func NewSQLite(db *sql.DB) *SQLite {
	return &SQLite{
		db: db,
	}
}

func (s *SQLite) Register(user *models.User) error {
	query := "INSERT INTO users VALUES (?, ?, ?, ?, ?, ?)"
	_, err := s.db.Exec(query, user.ID, user.Name, user.Email, user.Password, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: users.email" {
			return ErrEmailExists
		}
		return err
	}
	return nil
}

func (s *SQLite) GetUserByEmail(email string) (*models.User, error) {
	return s.getUser("SELECT * FROM users WHERE email = ?", email)
}

func (s *SQLite) GetUserByID(id uuid.UUID) (*models.User, error) {
	return s.getUser("SELECT * FROM users WHERE id = ?", id)
}

func (s *SQLite) getUser(query string, args ...any) (*models.User, error) {
	row := s.db.QueryRow(query, args...)

	user := new(models.User)
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *SQLite) CreateTables() {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS users(
		id UUID PRIMARY KEY NOT NULL,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
		)`)
	if err != nil {
		log.Fatal(err)
	}
}
