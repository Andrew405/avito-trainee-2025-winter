package main

import (
	"Avito-trainee/internal/auth"
	"Avito-trainee/internal/coin"
	"Avito-trainee/internal/info"
	"Avito-trainee/internal/merch"
	"log"

	"Avito-trainee/internal/config"
	"Avito-trainee/internal/db"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Can't load config: %v\n", err)
	}

	dbConn, err := db.NewPostgresConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Can't connect to database: %v\n", err)
	}
	defer dbConn.Close()

	err = db.RunMigrations(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	authService := auth.NewAuthService(dbConn, cfg.JWTSecret)
	coinService := coin.NewCoinService(dbConn)
	merchService := merch.NewMerchService(dbConn)
	infoService := info.NewInfoService(dbConn)

	_, _, _, _ = authService, coinService, merchService, infoService

	// TODO: init router

	// TODO: init server
}
