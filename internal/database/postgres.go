package database

import (
	"database/sql"
	"fmt"
	"github.com/datasektionen/GOrdian/internal/config"
	_ "github.com/lib/pq"
)

func Connect() (*sql.DB, error) {

	envVar := config.GetEnv()

	psqlconnString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		envVar.DBHost, envVar.DBPort, envVar.DBUser, envVar.DBPass, envVar.DBName)

	db, err := sql.Open("postgres", psqlconnString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	fmt.Println("Connected to database")

	return db, nil
}

func Close(db *sql.DB) error {
	err := db.Close()
	if err != nil {
		return fmt.Errorf("error closing connection to database: %v", err)
	}
	return nil
}
