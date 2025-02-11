package config

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	Environment string
	Port        string
	DatabaseURL string
	JWTSecret   string
	ReadTimeout int
}

func LoadConfig() (*Config, error) {
	conf := &Config{}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	conf.Environment = env

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	conf.Port = port

	dURL := os.Getenv("DATABASE_URL")
	if dURL == "" {
		return nil, errors.New("DATABASE_URL is not set")
	}
	conf.DatabaseURL = dURL

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET is not set")
	}
	conf.JWTSecret = jwtSecret

	if readTimeoutStr := os.Getenv("READ_TIMEOUT"); readTimeoutStr != "" {
		if rt, err := strconv.Atoi(readTimeoutStr); err == nil {
			conf.ReadTimeout = rt
		} else {
			return nil, errors.New("invalid READ_TIMEOUT value")
		}
	} else {
		conf.ReadTimeout = 5
	}

	return conf, nil
}
