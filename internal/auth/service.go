package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Authenticate(ctx context.Context, username, password string) (string, error)
}

type service struct {
	db         *sql.DB
	jwtManager *JWTManager
}

func NewAuthService(db *sql.DB, jwtSecret string) Service {
	jwtManager := NewJWTManager(jwtSecret, 24*60*60*time.Second)
	return &service{
		db:         db,
		jwtManager: jwtManager,
	}
}

func (s *service) Authenticate(ctx context.Context, username, password string) (string, error) {
	var storedHash string
	err := s.db.QueryRowContext(
		ctx,
		`SELECT password_hash FROM users WHERE username = $1`,
		username,
	).Scan(&storedHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return "", err
			}

			_, err = s.db.ExecContext(
				ctx,
				"INSERT INTO users (username, password_hash, coins) VALUES ($1, $2, $3)",
				username,
				string(hashedPwd),
				1000,
			)
			if err != nil {
				return "", err
			}
			return s.jwtManager.Generate(username)
		}
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}
	return s.jwtManager.Generate(username)
}
