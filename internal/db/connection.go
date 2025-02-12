package db

import (
	"database/sql"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

func RunMigrations(databaseURL string) error {
	m, err := migrate.New("file://internal/db/migrations", databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	// Применяем миграции
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			// Если нет изменений, это нормально
			return nil
		}
		return err
	}
	return nil
}
