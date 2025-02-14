package merch

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
)

var (
	ErrItemNotFound      = errors.New("can't find the product")
	ErrInsufficientCoins = errors.New("not enough coins")
)

var itemPrices = map[string]int{
	"t-shirt":    80,
	"cup":        20,
	"book":       50,
	"pen":        10,
	"powerbank":  200,
	"hoody":      300,
	"umbrella":   200,
	"socks":      10,
	"wallet":     50,
	"pink-hoody": 500,
}

type Service interface {
	BuyItem(ctx context.Context, userID int, item string) error
}

type service struct {
	db *sql.DB
}

func NewMerchService(db *sql.DB) Service {
	return &service{db: db}
}
func (s *service) BuyItem(ctx context.Context, userID int, item string) error {
	const op = "merch/service/BuyItem"
	price, ok := itemPrices[item]
	if !ok {
		return ErrItemNotFound
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return fmt.Errorf("%v: unable to start transaction: %w", op, err)
	}
	defer tx.Rollback()

	var coins int
	err = tx.QueryRowContext(
		ctx,
		`SELECT coins FROM users WHERE id = $1`,
		userID,
	).Scan(&coins)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("User not found: %v", err)
			return fmt.Errorf("%v: unable to find user: %v: %w", op, userID, err)
		}
		log.Printf("Unable to read balance: %v", err)
		return fmt.Errorf("%v: unable to read balance: %w", op, err)
	}

	if coins < price {
		log.Printf("Not enough coins")
		return ErrInsufficientCoins
	}

	// Списание монет
	_, err = tx.ExecContext(
		ctx,
		`UPDATE users SET coins = coins - $1 WHERE id = $2`,
		price,
		userID,
	)
	if err != nil {
		log.Printf("Debit error: %v", err)
		return fmt.Errorf("%v: debit error: %w", op, err)
	}

	// Запись транзакции
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO transactions (user_id, type, merch, amount) 
				VALUES ($1, $2, $3, $4);`,
		userID,
		"purchased",
		item,
		price,
	)
	if err != nil {
		log.Printf("Transaction error: %v", err)
		return fmt.Errorf("%v: transaction error: %w", op, err)
	}

	// Добавление в инвентарь
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO inventory (user_id, item_type, quantity) 
				VALUES ($1, $2, 1) 
				ON CONFLICT (user_id, item_type) 
    			DO UPDATE SET quantity = inventory.quantity + 1`,
		userID,
		item,
	)
	if err != nil {
		log.Printf("Inventory error: %v", err)
		return fmt.Errorf("%v: unable to update inventory: %w", op, err)
	}

	return tx.Commit()
}
