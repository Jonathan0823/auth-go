package utils

import "database/sql"

func AutoMigrate(db *sql.DB) error {
	if _, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS users ( 
      id SERIAL PRIMARY KEY,
      oauth_id VARCHAR(100) UNIQUE,
      username VARCHAR(100) NOT NULL,
      avatar_url VARCHAR(255),
      email VARCHAR(100) NOT NULL UNIQUE,
      password VARCHAR(100) NOT NULL,
      is_verified BOOLEAN DEFAULT FALSE,
      provider VARCHAR(50) DEFAULT 'local',
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`); err != nil {
		return err
	}

	if _, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS verify_emails ( 
      id UUID PRIMARY KEY, 
      user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      email VARCHAR(100) NOT NULL,
      expired_at TIMESTAMP NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`); err != nil {
		return err
	}

	if _, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS forgot_password_emails ( 
      id UUID PRIMARY KEY, 
      user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      email VARCHAR(100) NOT NULL,
      expired_at TIMESTAMP NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`); err != nil {
		return err
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS token_log ( 
			id UUID PRIMARY KEY,
			user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			jti VARCHAR(100) NOT NULL UNIQUE,
		  refreshed_from_jti VARCHAR(100),
		  invalidated_at TIMESTAMP,
			expired_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			ip_address VARCHAR(45) NOT NULL,
			user_agent TEXT NOT NULL,
			FOREIGN KEY (refreshed_from_jti) REFERENCES token_log(jti)
		);

		CREATE INDEX IF NOT EXISTS idx_token_log_user_id ON token_log(user_id);
		CREATE INDEX IF NOT EXISTS idx_token_log_jti ON token_log(jti);
		`); err != nil {
		return err
	}

	return nil
}
