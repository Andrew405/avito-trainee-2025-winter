package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

// JWTManager Структура менеджера:
//
//	Секретный ключ
//	Длительность токена
type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

// NewJWTManager Функция создания нового менеджера
func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

// Generate Генерация токена по имени пользователя
func (j *JWTManager) Generate(username string) (string, error) {
	claims := jwt.StandardClaims{
		Subject:   username,
		ExpiresAt: time.Now().Add(j.tokenDuration).Unix(),
		IssuedAt:  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// Verify Подтверждение токена и возврат имени пользователя
func (j *JWTManager) Verify(accessToken string) (string, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&jwt.StandardClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(j.secretKey), nil
		},
	)
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token")
	}

	return claims.Subject, nil
}
