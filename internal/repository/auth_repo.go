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

func (r *repository) VerifyEmail(token string) error {
	if _, err := r.db.Exec("UPDATE users SET is_verified = true WHERE email = (SELECT email FROM verify_emails WHERE id = $1)", token); err != nil {
		return err
	}

	if _, err := r.db.Exec("DELETE FROM verify_emails WHERE id = $1", token); err != nil {
		return err
	}

	return nil
}
