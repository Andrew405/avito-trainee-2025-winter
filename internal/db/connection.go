package db

import (
	"database/sql"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"log"
	"os"
	"path/filepath"
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

	dbConn.SetMaxOpenConns(100)                // Максимальное количество открытых соединений
	dbConn.SetMaxIdleConns(100)                // Максимальное количество простаивающих соединений
	dbConn.SetConnMaxLifetime(5 * time.Minute) // Максимальное время жизни соединения

	if err := dbConn.Ping(); err != nil {
		return nil, err
	}

	return dbConn, nil
}

func RunMigrations(databaseURL string) error {
	// Получаем абсолютный путь к корневой директории проекта
	rootDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// Формируем путь к миграциям
	migrationsPath := filepath.Join(rootDir, "internal", "db", "migrations")

	// Создаем источник миграций
	sourceURL := "file://" + migrationsPath

	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		log.Printf("Failed to initialize migrations: %v", err)
		return err
	}
	defer m.Close()

	// Применяем миграции
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			// Если нет изменений, это нормально
			log.Println("No new migrations to apply")
			return nil
		}
		log.Printf("Migration failed: %v", err)
		return err
	}
	log.Println("Migrations applied successfully")
	return nil
}
