package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"Avito-trainee/internal/auth"
	"Avito-trainee/internal/config"
	"Avito-trainee/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestSendCoinIntegration(t *testing.T) {
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
	_, err = dbConn.Exec(`TRUNCATE TABLE users, transactions RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	// Создание сервиса аутентификации
	authService := auth.NewAuthService(dbConn, cfg.JWTSecret)

	// Регистрация первого пользователя (отправитель)
	username1 := "user1"
	password1 := "password123"
	token1, err := authService.Authenticate(context.Background(), username1, password1)
	if err != nil {
		t.Fatalf("failed to authenticate user1: %v", err)
	}

	// Регистрация второго пользователя (получатель)
	username2 := "user2"
	password2 := "password123"
	_, err = authService.Authenticate(context.Background(), username2, password2)
	if err != nil {
		t.Fatalf("failed to authenticate user2: %v", err)
	}

	// Создание HTTP-сервера для тестирования
	router := setupRouter(dbConn, cfg.JWTSecret)

	// Тестовый запрос на отправку монет
	reqBody := map[string]interface{}{
		"toUser": username2,
		"amount": 500,
	}
	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/api/sendCoin", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token1)
	req.Header.Set("Content-Type", "application/json")

	// Выполнение запроса
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Проверка статуса ответа
	assert.Equal(t, http.StatusOK, rr.Code)

	// Проверка баланса отправителя
	var senderBalance int
	err = dbConn.QueryRow(`SELECT coins FROM users WHERE username = $1`, username1).Scan(&senderBalance)
	if err != nil {
		t.Fatalf("failed to fetch sender balance: %v", err)
	}
	assert.Equal(t, 500, senderBalance, "sender balance should be reduced by the sent amount")

	// Проверка баланса получателя
	var recipientBalance int
	err = dbConn.QueryRow(`SELECT coins FROM users WHERE username = $1`, username2).Scan(&recipientBalance)
	if err != nil {
		t.Fatalf("failed to fetch recipient balance: %v", err)
	}
	assert.Equal(t, 1500, recipientBalance, "recipient balance should be increased by the received amount")

	// Проверка записи транзакции
	var transactionCount int
	err = dbConn.QueryRow(`SELECT COUNT(*) FROM transactions WHERE user_id = (SELECT id FROM users WHERE username = $1) AND counterparty = $2`, username1, username2).Scan(&transactionCount)
	if err != nil {
		t.Fatalf("failed to fetch transactions: %v", err)
	}
	assert.Equal(t, 1, transactionCount, "there should be one transaction for the transfer")
}
