package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"Avito-trainee/internal/auth"
	"Avito-trainee/internal/config"
	"Avito-trainee/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthIntegration(t *testing.T) {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	assert.NoError(t, err)

	// Подключение к базе данных
	dbConn, err := db.NewPostgresConnection(cfg.DatabaseURL)
	assert.NoError(t, err)
	defer dbConn.Close()

	// Инициализация сервиса аутентификации
	authService := auth.NewAuthService(dbConn, cfg.JWTSecret)

	// Создание роутера
	r := chi.NewRouter()
	r.Post("/api/auth", auth.MakeAuthHandler(authService))

	// Создание тестового сервера
	testServer := httptest.NewServer(r)
	defer testServer.Close()

	// Тестовые данные
	payload := map[string]string{
		"username": "testuser",
		"password": "testpassword",
	}
	jsonPayload, _ := json.Marshal(payload)

	// Отправка POST-запроса на /api/auth
	resp, err := http.Post(testServer.URL+"/api/auth", "application/json", bytes.NewBuffer(jsonPayload))
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Проверка статуса ответа
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Проверка содержимого ответа
	var authResponse map[string]string
	err = json.NewDecoder(resp.Body).Decode(&authResponse)
	assert.NoError(t, err)
	assert.NotEmpty(t, authResponse["token"], "JWT token should be present in the response")
}
