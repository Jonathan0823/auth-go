package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Jonathan0823/auth-go/internal/models"
)

func (r *repository) GetUserByID(id int) (models.User, error) {
	var user models.User
	query := "SELECT id, username, email, updated_at, created_at FROM users WHERE id = $1"
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Email, &user.UpdatedAt, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, fmt.Errorf("user with id %d is not found", id)
		}
		return user, err
	}
	return user, nil
}

func (r *repository) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	query := "SELECT id, username, email, updated_at, created_at FROM users WHERE email = $1"
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Email, &user.UpdatedAt, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, fmt.Errorf("user with email %s is not found", email)
		}
		return user, err
	}
	return user, nil
}

func (r *repository) CreateUser(user models.User) error {
	query := "INSERT INTO users (username, email, password) VALUES ($1, $2, $3)"
	_, err := r.db.Exec(query, user.Username, user.Email, user.Password)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) GetAllUsers() ([]models.User, error) {
	var users []models.User
	query := "SELECT id, username, email, updated_at, created_at FROM users"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.UpdatedAt, &user.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %v", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *repository) UpdateUser(user models.UpdateUserRequest) error {
	query := "UPDATE users SET username = $1, email = $2 WHERE id = $3"
	_, err := r.db.Exec(query, user.Username, user.Email, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}
	return nil
}

func (r *repository) DeleteUser(id int) error {
	query := "DELETE FROM users WHERE id = $1"
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}
	return nil
}

func (r *repository) UpdateUserPassword(id int, newPassword string) error {
	query := "UPDATE users SET password = $1 WHERE id = $2"
	_, err := r.db.Exec(query, newPassword, id)
	if err != nil {
		return fmt.Errorf("failed to update user password: %v", err)
	}
	return nil
}

func (r *repository) GetPasswordByEmail(email string) (string, error) {
	var password string
	query := "SELECT password FROM users WHERE email = $1"
	err := r.db.QueryRow(query, email).Scan(&password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("user with email %s is not found", email)
		}
		return "", fmt.Errorf("failed to get password: %v", err)
	}
	return password, nil
}
