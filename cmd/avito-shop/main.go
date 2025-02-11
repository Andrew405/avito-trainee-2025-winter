package main

import (
	"Avito-trainee/internal/config"
	"Avito-trainee/internal/db"
	"log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Can't load config %v\n", err)
	}

	dbConn, err := db.NewPostgresConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Can't connect to database: %v\n", err)
	}
	defer dbConn.Close()

	// TODO: init services

	// TODO: init router

	// TODO: init server
}
