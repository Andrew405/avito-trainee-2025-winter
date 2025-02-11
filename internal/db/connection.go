package db

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

func NewPostgresConnection(databaseURL string) (*sql.DB, error) {
	dbConn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	dbConn.SetMaxOpenConns(25)                 // Максимальное количество открытых соединений
	dbConn.SetMaxIdleConns(25)                 // Максимальное количество простаивающих соединений
	dbConn.SetConnMaxLifetime(5 * time.Minute) // Максимальное время жизни соединения

	if err := dbConn.Ping(); err != nil {
		return nil, err
	}

	return dbConn, nil
}
