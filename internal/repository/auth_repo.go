// Package repository provides methods for interacting with the database related to email verification and password reset functionalities
package repository

import "github.com/Jonathan0823/auth-go/internal/models"

func (r *repository) CreateVerifyEmail(verifyEmail models.VerifyEmail) error {
	_, err := r.db.Exec("INSERT INTO verify_emails (id, user_id, email, expired_at) VALUES ($1, $2, $3, $4)", verifyEmail.ID, verifyEmail.UserID, verifyEmail.Email, verifyEmail.ExpiredAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) GetVerifyEmailByID(id string) (models.VerifyEmail, error) {
	var verifyEmail models.VerifyEmail
	err := r.db.QueryRow("SELECT id, email, expired_at FROM verify_emails WHERE id = $1", id).Scan(&verifyEmail.ID, &verifyEmail.Email, &verifyEmail.ExpiredAt)
	if err != nil {
		return models.VerifyEmail{}, err
	}
	return verifyEmail, nil
}

func (r *repository) VerifyEmail(id string) error {
	if _, err := r.db.Exec("UPDATE users SET is_verified = true WHERE email = (SELECT email FROM verify_emails WHERE id = $1)", id); err != nil {
		return err
	}

	if _, err := r.db.Exec("DELETE FROM verify_emails WHERE id = $1", id); err != nil {
		return err
	}

	return nil
}

func (r *repository) CreateForgotPasswordEmail(data models.ForgotPassword) error {
	_, err := r.db.Exec("INSERT INTO forgot_password_emails (id, email, expired_at) VALUES ($1, $2, $3)", data.ID, data.Email, data.ExpiredAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) GetForgotPasswordByID(id string) (models.ForgotPassword, error) {
	var data models.ForgotPassword
	err := r.db.QueryRow("SELECT id, email, expired_at FROM forgot_password_emails WHERE id = $1", id).Scan(&data.ID, &data.Email, &data.ExpiredAt)
	if err != nil {
		return models.ForgotPassword{}, err
	}
	return data, nil
}

func (r *repository) DeleteForgotPasswordByID(id string) error {
	_, err := r.db.Exec("DELETE FROM forgot_password_emails WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) CreateTokenLog(tokenLog models.TokenLog) error {
	_, err := r.db.Exec("INSERT INTO token_log (id, user_id, jti, refreshed_from_jti, invalidated_at, expired_at, created_at, ip_address, user_agent) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		tokenLog.ID, tokenLog.UserID, tokenLog.JTI, tokenLog.RefreshedFromJTI, tokenLog.InvalidatedAt, tokenLog.ExpiredAt, tokenLog.CreatedAt, tokenLog.IPAddress, tokenLog.UserAgent)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) GetTokenLogByJTI(jti string) (models.TokenLog, error) {
	var tokenLog models.TokenLog
	err := r.db.QueryRow("SELECT id, user_id, jti, refreshed_from_jti, invalidated_at, expired_at, created_at, ip_address, user_agent FROM token_log WHERE jti = $1", jti).Scan(
		&tokenLog.ID, &tokenLog.UserID, &tokenLog.JTI, &tokenLog.RefreshedFromJTI, &tokenLog.InvalidatedAt, &tokenLog.ExpiredAt, &tokenLog.CreatedAt, &tokenLog.IPAddress, &tokenLog.UserAgent)
	if err != nil {
		return models.TokenLog{}, err
	}
	return tokenLog, nil
}

func (r *repository) InvalidateTokenLog(oldJTI, newJTI string) error {
	_, err := r.db.Exec(`
		UPDATE
		token_log 
		SET invalidated_at = NOW(), refreshed_from_jti = $1
		WHERE jti = $2`, newJTI, oldJTI)
	if err != nil {
		return err
	}
	return nil
}
