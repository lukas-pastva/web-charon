package models

import (
	"database/sql"
	"fmt"
	"time"
)

type User struct {
	ID           int64
	Name         string
	Surname      string
	Nickname     string
	PasswordHash string
	IsAdmin      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserStore struct {
	DB *sql.DB
}

func (s *UserStore) GetAll() ([]User, error) {
	rows, err := s.DB.Query("SELECT id, name, surname, nickname, password_hash, is_admin, created_at, updated_at FROM users ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUsers(rows)
}

func (s *UserStore) GetByID(id int64) (*User, error) {
	u := &User{}
	err := s.DB.QueryRow("SELECT id, name, surname, nickname, password_hash, is_admin, created_at, updated_at FROM users WHERE id = ?", id).
		Scan(&u.ID, &u.Name, &u.Surname, &u.Nickname, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserStore) GetByNickname(nickname string) (*User, error) {
	u := &User{}
	err := s.DB.QueryRow("SELECT id, name, surname, nickname, password_hash, is_admin, created_at, updated_at FROM users WHERE nickname = ?", nickname).
		Scan(&u.ID, &u.Name, &u.Surname, &u.Nickname, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserStore) Create(u *User) error {
	res, err := s.DB.Exec("INSERT INTO users (name, surname, nickname, password_hash, is_admin) VALUES (?, ?, ?, ?, ?)",
		u.Name, u.Surname, u.Nickname, u.PasswordHash, u.IsAdmin)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	u.ID, _ = res.LastInsertId()
	return nil
}

func (s *UserStore) Update(u *User) error {
	_, err := s.DB.Exec("UPDATE users SET name=?, surname=?, nickname=?, password_hash=?, is_admin=? WHERE id=?",
		u.Name, u.Surname, u.Nickname, u.PasswordHash, u.IsAdmin, u.ID)
	return err
}

func (s *UserStore) Delete(id int64) error {
	_, err := s.DB.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

func (s *UserStore) GetFirstAdminID() (int64, error) {
	var id int64
	err := s.DB.QueryRow("SELECT id FROM users WHERE is_admin = TRUE ORDER BY id ASC LIMIT 1").Scan(&id)
	return id, err
}

func (s *UserStore) Count() (int, error) {
	var count int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func scanUsers(rows *sql.Rows) ([]User, error) {
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Surname, &u.Nickname, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
