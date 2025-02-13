package coin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrInsufficientFunds = errors.New("not enough coins")
	ErrSameUser          = errors.New("unable to send coins to yourself")
)

type Service interface {
	SendCoin(ctx context.Context, fromUserID int, toUsername string, amount int) error
}

type service struct {
	db *sql.DB
}

func NewCoinService(db *sql.DB) Service {
	return &service{db: db}
}

func (s *service) SendCoin(ctx context.Context, fromUserID int, toUsername string, amount int) error {
	const op = "coin/service/SendCoin"
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%v: unable to start transaction: %w", op, err)
	}
	defer tx.Rollback()

	var toUserID int
	err = tx.QueryRowContext(
		ctx,
		"SELECT id FROM users WHERE username = $1",
		toUsername,
	).Scan(&toUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%v: user not found: %s", op, toUsername)
		}
		return fmt.Errorf("%v: unable to find user: %w", op, err)
	}

	var fromUsername string
	err = tx.QueryRowContext(
		ctx,
		"SELECT username FROM users WHERE id = $1",
		fromUserID,
	).Scan(&fromUsername)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%v: user id not found: %v", op, fromUserID)
		}
		return fmt.Errorf("%v: unable to find user: %w", op, err)
	}

	if fromUserID == toUserID {
		return ErrSameUser
	}

	var currentBalance int
	err = tx.QueryRowContext(
		ctx,
		"SELECT coins FROM users WHERE id = $1 FOR UPDATE ",
		fromUserID,
	).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("%v: unable to check balance: %w", op, err)
	}

	if currentBalance < amount {
		return ErrInsufficientFunds
	}

	_, err = tx.ExecContext(
		ctx,
		"UPDATE users SET coins = coins - $1 WHERE id = $2",
		amount,
		fromUserID,
	)
	if err != nil {
		return fmt.Errorf("%v: debit error: %w", op, err)
	}

	_, err = tx.ExecContext(
		ctx,
		"UPDATE users SET coins = coins + $1 WHERE id = $2",
		amount,
		toUserID,
	)
	if err != nil {
		return fmt.Errorf("%v: credit error: %w", op, err)
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO transactions (user_id, type, counterparty, amount) 
				VALUES ($1, 'sent', $2, $3), ($4, 'received', $5, $6)`,
		fromUserID,
		toUsername,
		amount,
		toUserID,
		fromUsername,
		amount,
	)
	if err != nil {
		return fmt.Errorf("%v: transaction error: %w", op, err)
	}

	return tx.Commit()
}
