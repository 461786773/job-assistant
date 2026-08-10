package db

import (
	"database/sql"
	"fmt"
	"strings"
)

type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	CreatedAt    string `json:"createdAt"`
}

func (s *Store) CreateUser(u *User) error {
	_, err := s.db.Exec(`
INSERT INTO users (id, username, password_hash, created_at)
VALUES (?, ?, ?, ?)`,
		u.ID, u.Username, u.PasswordHash, u.CreatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("用户名已存在")
		}
		return err
	}
	return nil
}

func (s *Store) GetUserByUsername(username string) (*User, error) {
	row := s.db.QueryRow(`
SELECT id, username, password_hash, created_at FROM users WHERE username = ?`, username)
	return scanUser(row)
}

func (s *Store) GetUserByID(id string) (*User, error) {
	row := s.db.QueryRow(`
SELECT id, username, password_hash, created_at FROM users WHERE id = ?`, id)
	return scanUser(row)
}

func scanUser(row *sql.Row) (*User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
