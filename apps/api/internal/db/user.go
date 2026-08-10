package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
)

const DefaultUsername = "default"
const DefaultUserID = "default"

type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	PrimaryNeed  string `json:"primaryNeed,omitempty"`
	CreatedAt    string `json:"createdAt"`
}

func newUserID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Store) CreateUser(u *User) error {
	_, err := s.db.Exec(`
INSERT INTO users (id, username, password_hash, primary_need, created_at)
VALUES (?, ?, ?, ?, ?)`,
		u.ID, u.Username, u.PasswordHash, u.PrimaryNeed, u.CreatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("用户名已存在")
		}
		return err
	}
	return nil
}

// EnsureUser returns existing user by username, or creates one (no password).
func (s *Store) EnsureUser(username string) (*User, error) {
	u, err := s.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	if u != nil {
		return u, nil
	}
	u = &User{
		ID:           newUserID(),
		Username:     username,
		PasswordHash: "",
		CreatedAt:    Now(),
	}
	if username == DefaultUsername {
		u.ID = DefaultUserID
	}
	if err := s.CreateUser(u); err != nil {
		if strings.Contains(err.Error(), "用户名已存在") {
			return s.GetUserByUsername(username)
		}
		return nil, err
	}
	return u, nil
}

func (s *Store) seedDefaultUser() error {
	if _, err := s.EnsureUser(DefaultUsername); err != nil {
		return err
	}
	_, err := s.db.Exec(`UPDATE tasks SET user_id = ? WHERE user_id = '' OR user_id IS NULL`, DefaultUserID)
	return err
}

func (s *Store) GetUserByUsername(username string) (*User, error) {
	row := s.db.QueryRow(`
SELECT id, username, password_hash, COALESCE(primary_need, ''), created_at FROM users WHERE username = ?`, username)
	return scanUser(row)
}

func (s *Store) GetUserByID(id string) (*User, error) {
	row := s.db.QueryRow(`
SELECT id, username, password_hash, COALESCE(primary_need, ''), created_at FROM users WHERE id = ?`, id)
	return scanUser(row)
}

func (s *Store) UpdateUserPrimaryNeed(userID, need string) error {
	res, err := s.db.Exec(`UPDATE users SET primary_need = ? WHERE id = ?`, need, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func scanUser(row *sql.Row) (*User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.PrimaryNeed, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
