package config

import "database/sql"

func AutoMigrate(db *sql.DB) error {
	if _, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS users ( 
      id SERIAL PRIMARY KEY,
      username VARCHAR(100) NOT NULL,
      email VARCHAR(100) NOT NULL UNIQUE,
      password VARCHAR(100) NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`); err != nil {
		return err
	}

	if _, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS verify_emails ( 
      id UUID PRIMARY KEY, 
      email VARCHAR(100) NOT NULL,
      expired_at TIMESTAMP NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    )`); err != nil {
		return err
	}
	return nil
}
