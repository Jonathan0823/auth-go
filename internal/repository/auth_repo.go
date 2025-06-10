package repository

import "github.com/Jonathan0823/auth-go/internal/models"

func (r *repository) CreateVerifyEmail(verifyEmail models.VerifyEmail) error {
	_, err := r.db.Exec("INSERT INTO verify_emails (id, email, expired_at) VALUES ($1, $2, $3)", verifyEmail.ID, verifyEmail.Email, verifyEmail.ExpiredAt)
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
