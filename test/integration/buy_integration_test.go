package integration

import (
	"Avito-trainee/internal/coin"
	"Avito-trainee/internal/config"
	middleware2 "Avito-trainee/internal/middleware"
	"bytes"
	"context"
	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"net/http/httptest"
	"testing"

	"Avito-trainee/internal/auth"
	"Avito-trainee/internal/db"
	"Avito-trainee/internal/merch"
	"github.com/stretchr/testify/assert"
)

func TestBuyItemIntegration(t *testing.T) {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Инициализация подключения к базе данных
	dbConn, err := db.NewPostgresConnection(cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	// Применение миграций
	if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Очистка базы данных перед тестом
	_, err = dbConn.Exec(`TRUNCATE TABLE users, transactions, inventory RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	// Создание сервиса аутентификации
	authService := auth.NewAuthService(dbConn, cfg.JWTSecret)

	// Регистрация пользователя
	username := "testuser"
	password := "password123"
	token, err := authService.Authenticate(context.Background(), username, password)
	if err != nil {
		t.Fatalf("failed to authenticate user: %v", err)
	}

	// Создание HTTP-сервера для тестирования
	router := setupRouter(dbConn, cfg.JWTSecret)

	// Тестовый запрос на покупку товара
	item := "t-shirt"
	reqBody := bytes.NewBuffer(nil)
	req, err := http.NewRequest("GET", "/api/buy/"+item, reqBody)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// Выполнение запроса
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Проверка статуса ответа
	assert.Equal(t, http.StatusOK, rr.Code)

	// Проверка баланса пользователя
	var coins int
	err = dbConn.QueryRow(`SELECT coins FROM users WHERE username = $1`, username).Scan(&coins)
	if err != nil {
		t.Fatalf("failed to fetch user balance: %v", err)
	}
	assert.Equal(t, 920, coins, "balance should be reduced by the price of the item")

	// Проверка инвентаря пользователя
	var quantity int
	err = dbConn.QueryRow(`SELECT quantity FROM inventory WHERE user_id = (SELECT id FROM users WHERE username = $1) AND item_type = $2`, username, item).Scan(&quantity)
	if err != nil {
		t.Fatalf("failed to fetch inventory: %v", err)
	}
	assert.Equal(t, 1, quantity, "inventory should contain one item of the purchased type")

	// Проверка записи транзакции
	var transactionCount int
	err = dbConn.QueryRow(`SELECT COUNT(*) FROM transactions WHERE user_id = (SELECT id FROM users WHERE username = $1) AND merch = $2`, username, item).Scan(&transactionCount)
	if err != nil {
		t.Fatalf("failed to fetch transactions: %v", err)
	}
	assert.Equal(t, 1, transactionCount, "there should be one transaction for the purchase")
}

// setupRouter создает роутер с необходимыми обработчиками
func setupRouter(dbConn *sql.DB, jwtSecret string) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Инициализация сервисов
	authService := auth.NewAuthService(dbConn, jwtSecret)
	merchService := merch.NewMerchService(dbConn)
	coinService := coin.NewCoinService(dbConn)

	// Маршруты
	r.Route("/api", func(r chi.Router) {
		r.Post("/auth", auth.MakeAuthHandler(authService))
	})
	r.Group(func(r chi.Router) {
		r.Use(middleware2.JWTAuthMiddleware(dbConn, jwtSecret))
		r.Get("/api/buy/{item}", merch.MakeBuyHandler(merchService))
		r.Post("/api/sendCoin", coin.MakeSendCoinHandler(coinService))
	})

	return r
}
