package unit

import (
	"context"
	"database/sql"
	"testing"

	"Avito-trainee/internal/coin"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSendCoin_Success(t *testing.T) {
	// Создаем mock базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Инициализируем сервис
	service := coin.NewCoinService(db)

	// Настройка mock-запросов
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id FROM users WHERE username = \$1`).
		WithArgs("user2").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectQuery(`SELECT username FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("user1"))
	mock.ExpectQuery(`SELECT coins FROM users WHERE id = \$1 FOR UPDATE`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"coins"}).AddRow(500))
	mock.ExpectExec(`UPDATE users SET coins = coins \- \$1 WHERE id = \$2`).
		WithArgs(100, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET coins = coins \+ \$1 WHERE id = \$2`).
		WithArgs(100, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO transactions \(user_id, type, counterparty, amount\) VALUES \(\$1, 'sent', \$2, \$3\), \(\$4, 'received', \$5, \$6\)`).
		WithArgs(1, "user2", 100, 2, "user1", 100).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Выполняем тест
	err = service.SendCoin(context.Background(), 1, "user2", 100)
	assert.NoError(t, err)

	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestSendCoin_InsufficientFunds(t *testing.T) {
	// Создаем mock базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Инициализируем сервис
	service := coin.NewCoinService(db)

	// Настройка mock-запросов
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id FROM users WHERE username = \$1`).
		WithArgs("user2").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectQuery(`SELECT username FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("user1"))
	mock.ExpectQuery(`SELECT coins FROM users WHERE id = \$1 FOR UPDATE`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"coins"}).AddRow(50))
	mock.ExpectRollback()

	// Выполняем тест
	err = service.SendCoin(context.Background(), 1, "user2", 100)
	assert.Error(t, err)
	assert.Equal(t, coin.ErrInsufficientFunds, err)

	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestSendCoin_SameUser(t *testing.T) {
	// Создаем mock базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Инициализируем сервис
	service := coin.NewCoinService(db)

	// Настройка mock-запросов
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id FROM users WHERE username = \$1`).
		WithArgs("user1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(`SELECT username FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("user1"))
	mock.ExpectRollback()

	// Выполняем тест
	err = service.SendCoin(context.Background(), 1, "user1", 100)
	assert.Error(t, err)
	assert.Equal(t, coin.ErrSameUser, err)

	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestSendCoin_UserNotFound(t *testing.T) {
	// Создаем mock базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Инициализируем сервис
	service := coin.NewCoinService(db)

	// Настройка mock-запросов
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id FROM users WHERE username = \$1`).
		WithArgs("nonexistent_user").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	// Выполняем тест
	err = service.SendCoin(context.Background(), 1, "nonexistent_user", 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")

	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}
