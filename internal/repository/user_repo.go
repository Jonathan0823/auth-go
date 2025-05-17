package repository

import (
	"database/sql"
	"errors"

	"github.com/Jonathan0823/auth-go/internal/dto"
)

func (r *repository) GetUserByID(id int) (dto.User, error) {
	var user dto.User
	query := "SELECT id, username, email, updated_at, created_at FROM users WHERE id = $1"
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Email, &user.UpdatedAt, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, nil
		}
		return user, err
	}
	return user, nil
}

func (r *repository) GetUserByEmail(email string) (dto.User, error) {
	var user dto.User
	query := "SELECT id, username, email, password, updated_at, created_at FROM users WHERE email = $1"
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.UpdatedAt, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, nil
		}
		return user, err
	}
	return user, nil
}

func (r *repository) CreateUser(user dto.User) error {
	query := "INSERT INTO users (username, email, password) VALUES ($1, $2, $3)"
	_, err := r.db.Exec(query, user.Username, user.Email, user.Password)
	if err != nil {
		return err
	}
	return nil
}
