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
	"Avito-trainee/internal/info"
	"Avito-trainee/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestInfoIntegration(t *testing.T) {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	assert.NoError(t, err)

	// Подключение к базе данных
	dbConn, err := db.NewPostgresConnection(cfg.DatabaseURL)
	assert.NoError(t, err)
	defer dbConn.Close()

	// Инициализация сервисов
	authService := auth.NewAuthService(dbConn, cfg.JWTSecret)
	infoService := info.NewInfoService(dbConn)

	// Создание роутера
	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth", auth.MakeAuthHandler(authService))
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuthMiddleware(dbConn, cfg.JWTSecret)) // Middleware для проверки JWT
		r.Get("/api/info", info.MakeInfoHandler(infoService))
	})

	// Создание тестового сервера
	testServer := httptest.NewServer(r)
	defer testServer.Close()

	// Авторизация пользователя
	authPayload := map[string]string{
		"username": "testuser",
		"password": "testpassword",
	}
	authJSON, _ := json.Marshal(authPayload)
	authResp, err := http.Post(testServer.URL+"/api/auth", "application/json", bytes.NewBuffer(authJSON))
	assert.NoError(t, err)
	defer authResp.Body.Close()

	// Получение JWT-токена
	var authResponse map[string]string
	err = json.NewDecoder(authResp.Body).Decode(&authResponse)
	assert.NoError(t, err)
	token := authResponse["token"]

	// Отправка GET-запроса на /api/info с JWT-токеном
	req, err := http.NewRequest("GET", testServer.URL+"/api/info", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Проверка статуса ответа
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Проверка содержимого ответа
	var infoResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&infoResponse)
	assert.NoError(t, err)
	assert.Contains(t, infoResponse, "coins", "Response should contain 'coins' field")
	assert.Contains(t, infoResponse, "inventory", "Response should contain 'inventory' field")
	assert.Contains(t, infoResponse, "coinHistory", "Response should contain 'coinHistory' field")
}
