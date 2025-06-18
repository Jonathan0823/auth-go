// Package config provides configuration and database initialization for the application
package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/Jonathan0823/auth-go/utils"
	_ "github.com/lib/pq"
)

type Config struct {
	Port           string
	AllowedOrigins string
}

func InitDB() *sql.DB {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSL")

	if host == "" || port == "" || user == "" || password == "" || dbname == "" {
		log.Fatal("Database connection parameters are not set in environment variables")
	}

	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	))
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Error connecting to the database:", err)
	}

	if err = utils.AutoMigrate(db); err != nil {
		log.Fatal("Error migrating the database:", err)
	}

	return db
}
