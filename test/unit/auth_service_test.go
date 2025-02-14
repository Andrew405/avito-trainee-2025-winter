package unit

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"Avito-trainee/internal/auth"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthenticate_SuccessfulLogin(t *testing.T) {
	// Создаем mock базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Инициализируем сервис
	service := auth.NewAuthService(db, "test-secret")

	// Настройка mock-запросов
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	mock.ExpectQuery(`SELECT password_hash FROM users WHERE username = \$1`).
		WithArgs("user1").
		WillReturnRows(sqlmock.NewRows([]string{"password_hash"}).AddRow(string(hashedPassword)))

	// Выполняем тест
	token, err := service.Authenticate(context.Background(), "user1", "password123")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAuthenticate_InvalidCredentials(t *testing.T) {
	// Создаем mock базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Инициализируем сервис
	service := auth.NewAuthService(db, "test-secret")

	// Настройка mock-запросов
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	mock.ExpectQuery(`SELECT password_hash FROM users WHERE username = \$1`).
		WithArgs("user1").
		WillReturnRows(sqlmock.NewRows([]string{"password_hash"}).AddRow(string(hashedPassword)))

	// Выполняем тест
	_, err = service.Authenticate(context.Background(), "user1", "wrongpassword")
	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())

	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAuthenticate_NewUserRegistration(t *testing.T) {
	// Создаем mock базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Инициализируем сервис
	service := auth.NewAuthService(db, "test-secret")

	// Настройка mock-запросов
	mock.ExpectQuery(`SELECT password_hash FROM users WHERE username = \$1`).
		WithArgs("newuser").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO users \(username, password_hash, coins\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs("newuser", sqlmock.AnyArg(), 1000).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Выполняем тест
	token, err := service.Authenticate(context.Background(), "newuser", "password123")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAuthenticate_DatabaseError(t *testing.T) {
	// Создаем mock базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Инициализируем сервис
	service := auth.NewAuthService(db, "test-secret")

	// Настройка mock-запросов
	mock.ExpectQuery(`SELECT password_hash FROM users WHERE username = \$1`).
		WithArgs("user1").
		WillReturnError(errors.New("database error"))

	// Выполняем тест
	_, err = service.Authenticate(context.Background(), "user1", "password123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")

	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}
