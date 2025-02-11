package main

import (
	"Avito-trainee/internal/config"
	"log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Can't load config %v\n", err)
	}

	// TODO: init db

	// TODO: init services

	// TODO: init router

	// TODO: init server
}
