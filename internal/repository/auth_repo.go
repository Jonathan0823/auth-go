package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Jonathan0823/auth-go/internal/models"
)

type AuthRepository interface {
	CreateVerifyEmail(ctx context.Context, verifyEmail models.VerifyEmail) error
	GetVerifyEmailByID(ctx context.Context, id string) (models.VerifyEmail, error)
	VerifyEmail(ctx context.Context, id string) error
	CreateForgotPasswordEmail(ctx context.Context, data models.ForgotPassword) error
	GetForgotPasswordByID(ctx context.Context, id string) (models.ForgotPassword, error)
	DeleteForgotPasswordByID(ctx context.Context, id string) error
	CreateTokenLog(ctx context.Context, tokenLog models.TokenLog) error
	GetTokenLogByJTI(ctx context.Context, jti string) (models.TokenLog, error)
	InvalidateTokenLog(ctx context.Context, oldJti, newJti string) error
	IsTokenLogInvalidated(ctx context.Context, jti string) (bool, error)
}

type authRepository struct {
	db DBTX
}

func NewAuthRepository(dbtx DBTX) AuthRepository {
	return &authRepository{
		db: dbtx,
	}
}

func (r *authRepository) CreateVerifyEmail(ctx context.Context, verifyEmail models.VerifyEmail) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO verify_emails (id, user_id, email, expired_at) VALUES ($1, $2, $3, $4)", verifyEmail.ID, verifyEmail.UserID, verifyEmail.Email, verifyEmail.ExpiredAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *authRepository) GetVerifyEmailByID(ctx context.Context, id string) (models.VerifyEmail, error) {
	var verifyEmail models.VerifyEmail
	err := r.db.QueryRowContext(ctx, "SELECT id, email, expired_at FROM verify_emails WHERE id = $1", id).Scan(&verifyEmail.ID, &verifyEmail.Email, &verifyEmail.ExpiredAt)
	if err != nil {
		return models.VerifyEmail{}, err
	}
	return verifyEmail, nil
}

func (r *authRepository) VerifyEmail(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, "UPDATE users SET is_verified = true WHERE email = (SELECT email FROM verify_emails WHERE id = $1)", id); err != nil {
		return err
	}

	if _, err := r.db.ExecContext(ctx, "DELETE FROM verify_emails WHERE id = $1", id); err != nil {
		return err
	}

	return nil
}

func (r *authRepository) CreateForgotPasswordEmail(ctx context.Context, data models.ForgotPassword) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO forgot_password_emails (id, email, expired_at) VALUES ($1, $2, $3)", data.ID, data.Email, data.ExpiredAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *authRepository) GetForgotPasswordByID(ctx context.Context, id string) (models.ForgotPassword, error) {
	var data models.ForgotPassword
	err := r.db.QueryRowContext(ctx, "SELECT id, email, expired_at FROM forgot_password_emails WHERE id = $1", id).Scan(&data.ID, &data.Email, &data.ExpiredAt)
	if err != nil {
		return models.ForgotPassword{}, err
	}
	return data, nil
}

func (r *authRepository) DeleteForgotPasswordByID(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM forgot_password_emails WHERE id = $1", id)
	if err != nil {
		return err
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *authRepository) CreateTokenLog(ctx context.Context, tokenLog models.TokenLog) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO token_log (id, user_id, jti, refreshed_from_jti, invalidated_at, expired_at, created_at, ip_address, user_agent) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		tokenLog.ID, tokenLog.UserID, tokenLog.JTI, tokenLog.RefreshedFromJTI, tokenLog.InvalidatedAt, tokenLog.ExpiredAt, tokenLog.CreatedAt, tokenLog.IPAddress, tokenLog.UserAgent)
	if err != nil {
		return err
	}
	return nil
}

func (r *authRepository) GetTokenLogByJTI(ctx context.Context, jti string) (models.TokenLog, error) {
	var tokenLog models.TokenLog
	err := r.db.QueryRowContext(ctx, "SELECT id, user_id, jti, refreshed_from_jti, invalidated_at, expired_at, created_at, ip_address, user_agent FROM token_log WHERE jti = $1", jti).Scan(
		&tokenLog.ID, &tokenLog.UserID, &tokenLog.JTI, &tokenLog.RefreshedFromJTI, &tokenLog.InvalidatedAt, &tokenLog.ExpiredAt, &tokenLog.CreatedAt, &tokenLog.IPAddress, &tokenLog.UserAgent)
	if err != nil {
		return models.TokenLog{}, err
	}
	return tokenLog, nil
}

func (r *authRepository) InvalidateTokenLog(ctx context.Context, oldJTI, newJTI string) error {
	setClauses := []string{"invalidated_at = NOW()"}
	args := []any{oldJTI}

	if newJTI != "" {
		setClauses = append(setClauses, fmt.Sprintf("refreshed_from_jti = $%d", len(args)+1))
		args = append(args, newJTI)
	}

	setClause := strings.Join(setClauses, ", ")

	query := fmt.Sprintf(`
	UPDATE your_table
	SET %s
	WHERE jti = $1
`, setClause)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (r *authRepository) IsTokenLogInvalidated(ctx context.Context, jti string) (bool, error) {
	var invalidated bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM token_log WHERE jti = $1 AND invalidated_at IS NOT NULL)", jti).Scan(&invalidated)
	if err != nil {
		return false, err
	}
	return invalidated, nil
}
