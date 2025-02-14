package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

func JWTAuthMiddleware(db *sql.DB, secretKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(secretKey), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			username := claims["sub"].(string)
			var userID int
			err = db.QueryRowContext(r.Context(), "SELECT id FROM users WHERE username = $1", username).Scan(&userID)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "userID", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
