package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Jonathan0823/auth-go/internal/models"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string, includePassword bool) (*models.User, error)
	CreateUser(ctx context.Context, user models.User) error
	GetAllUsers(ctx context.Context) ([]*models.User, error)
	UpdateUser(ctx context.Context, user models.UpdateUserRequest) error
	DeleteUser(ctx context.Context, id int) error
	UpdateUserPassword(ctx context.Context, id int, newPassword string) error
}

type userRepository struct {
	db DBTX
}

func NewUserRepository(dbtx DBTX) UserRepository {
	return &userRepository{db: dbtx}
}

func (r *userRepository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	query := "SELECT id, username, email, updated_at, created_at FROM users WHERE id = $1"
	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Username, &user.Email, &user.UpdatedAt, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string, includePassword bool) (*models.User, error) {
	var user models.User
	var scanFields []any
	scanFields = append(scanFields, &user.ID, &user.Username, &user.Email, &user.UpdatedAt, &user.CreatedAt)
	selectFields := "id, username, email, updated_at, created_at"
	if includePassword {
		selectFields += ", password"
		scanFields = append(scanFields, &user.Password)
	}
	query := fmt.Sprintf(`
		SELECT 
		%s
		FROM users 
		WHERE email = $1`, selectFields)
	err := r.db.QueryRowContext(ctx, query, email).Scan(scanFields...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) CreateUser(ctx context.Context, user models.User) error {
	query := "INSERT INTO users (username, email, password) VALUES ($1, $2, $3)"
	_, err := r.db.ExecContext(ctx, query, user.Username, user.Email, user.Password)
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	query := "SELECT id, username, email, updated_at, created_at FROM users"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		user := new(models.User)
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.UpdatedAt, &user.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %v", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user models.UpdateUserRequest) error {
	query := "UPDATE users SET username = $1, email = $2 WHERE id = $3"
	_, err := r.db.ExecContext(ctx, query, user.Username, user.Email, user.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) DeleteUser(ctx context.Context, id int) error {
	query := "DELETE FROM users WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) UpdateUserPassword(ctx context.Context, id int, newPassword string) error {
	query := "UPDATE users SET password = $1 WHERE id = $2"
	_, err := r.db.ExecContext(ctx, query, newPassword, id)
	if err != nil {
		return err
	}
	return nil
}
