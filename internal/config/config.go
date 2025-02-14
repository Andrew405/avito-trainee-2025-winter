package config

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Port        string
	DatabaseURL string
	JWTSecret   string
	ReadTimeout int
}

func LoadConfig() (*Config, error) {
	// Получаем текущую рабочую директорию
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get current working directory: %v", err)
	}

	// Формируем абсолютный путь к .env
	envPath := filepath.Join(currentDir, ".env")

	// Загрузка переменных окружения из .env файла
	if err := godotenv.Load(envPath); err != nil {
		if err := godotenv.Load("/app/.env"); err != nil {
			return nil, errors.New("failed to load .env file")
		}
	}

	conf := &Config{}

	// Определение среды выполнения (development, production и т.д.)
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	conf.Environment = env

	// Порт приложения
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // По умолчанию 8080
	}
	conf.Port = port

	// URL базы данных
	dURL := os.Getenv("DATABASE_URL")
	if dURL == "" {
		return nil, errors.New("DATABASE_URL is not set")
	}
	conf.DatabaseURL = dURL

	// Секретный ключ для JWT
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET is not set")
	}
	conf.JWTSecret = jwtSecret

	// Таймаут чтения запросов
	if readTimeoutStr := os.Getenv("READ_TIMEOUT"); readTimeoutStr != "" {
		if rt, err := strconv.Atoi(readTimeoutStr); err == nil {
			conf.ReadTimeout = rt
		} else {
			return nil, errors.New("invalid READ_TIMEOUT value")
		}
	} else {
		conf.ReadTimeout = 5 // По умолчанию 5 секунд
	}

	return conf, nil
}
