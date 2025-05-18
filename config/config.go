package config

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type Config struct {
	Port           string
	AllowedOrigins string
}

func InitDB() *sql.DB {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set in .env file")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Error connecting to the database:", err)
	}

	if err = AutoMigrate(db); err != nil {
		log.Fatal("Error migrating the database:", err)
	}

	return db
}
