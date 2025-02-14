package unit

import (
	"context"
	"errors"
	"testing"

	"Avito-trainee/internal/merch"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestBuyItem_Success(t *testing.T) {
	// Создаем mock базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Инициализируем сервис
	service := merch.NewMerchService(db)

	// Настройка mock-запросов
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT coins FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"coins"}).AddRow(1000))
	mock.ExpectExec(`UPDATE users SET coins = coins - \$1 WHERE id = \$2`).
		WithArgs(80, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO transactions \(user_id, type, merch, amount\) VALUES \(\$1, \$2, \$3, \$4\)`).
		WithArgs(1, "purchased", "t-shirt", 80).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO inventory \(user_id, item_type, quantity\) VALUES \(\$1, \$2, 1\) ON CONFLICT \(user_id, item_type\) DO UPDATE SET quantity = inventory.quantity \+ 1`).
		WithArgs(1, "t-shirt").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Выполняем тест
	err = service.BuyItem(context.Background(), 1, "t-shirt")
	assert.NoError(t, err)

	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestBuyItem_InsufficientCoins(t *testing.T) {
	// Создаем mock базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Инициализируем сервис
	service := merch.NewMerchService(db)

	// Настройка mock-запросов
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT coins FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"coins"}).AddRow(50))
	mock.ExpectRollback()

	// Выполняем тест
	err = service.BuyItem(context.Background(), 1, "t-shirt")
	assert.Error(t, err)
	assert.Equal(t, merch.ErrInsufficientCoins, err)

	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestBuyItem_ItemNotFound(t *testing.T) {
	// Создаем mock базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Инициализируем сервис
	service := merch.NewMerchService(db)

	// Выполняем тест
	err = service.BuyItem(context.Background(), 1, "nonexistent-item")
	assert.Error(t, err)
	assert.Equal(t, merch.ErrItemNotFound, err)

	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestBuyItem_DatabaseError(t *testing.T) {
	// Создаем mock базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Инициализируем сервис
	service := merch.NewMerchService(db)

	// Настройка mock-запросов
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT coins FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	// Выполняем тест
	err = service.BuyItem(context.Background(), 1, "t-shirt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")

	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}
