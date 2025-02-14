package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"Avito-trainee/internal/auth"
	"Avito-trainee/internal/coin"
	"Avito-trainee/internal/config"
	"Avito-trainee/internal/db"
	"Avito-trainee/internal/info"
	"Avito-trainee/internal/merch"
	middleware2 "Avito-trainee/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth", auth.MakeAuthHandler(authService))
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware2.JWTAuthMiddleware(dbConn, cfg.JWTSecret))

		r.Get("/api/info", info.MakeInfoHandler(infoService))
		r.Post("/api/sendCoin", coin.MakeSendCoinHandler(coinService))
		r.Get("/api/buy/{item}", merch.MakeBuyHandler(merchService))
	})

	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Server is listening on %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("could not listen on %s: %v", server.Addr, err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Could not gracefully shutdown the server: %v", err)
	}
	log.Println("Server stopped")
}
