package info

import (
	"context"
	"database/sql"
	"fmt"
)

type InfoResponse struct {
	Coins       int             `json:"coins"`
	Inventory   []InventoryItem `json:"inventory"`
	CoinHistory CoinHistory     `json:"coinHistory"`
}

type InventoryItem struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinHistory struct {
	Received []Transaction `json:"received"`
	Sent     []Transaction `json:"sent"`
}

type Transaction struct {
	FromUser string `json:"fromUser,omitempty"`
	ToUser   string `json:"toUser,omitempty"`
	Amount   int    `json:"amount"`
}

type Service interface {
	GetInfo(ctx context.Context, userID int) (InfoResponse, error)
}
type service struct {
	db *sql.DB
}

func NewInfoService(db *sql.DB) Service {
	return &service{db: db}
}

func (s *service) GetInfo(ctx context.Context, userID int) (InfoResponse, error) {
	const op = "info/service/GetInfo"
	var coins int
	err := s.db.QueryRowContext(
		ctx,
		"SELECT coins FROM users WHERE id = $1",
		userID,
	).Scan(&coins)
	if err != nil {
		return InfoResponse{}, fmt.Errorf("%v: unable to get balance: %w", err)
	}

	rows, err := s.db.QueryContext(
		ctx,
		"SELECT item_type, quantity FROM inventory WHERE user_id = $1",
		userID,
	)
	if err != nil {
		return InfoResponse{}, fmt.Errorf("%v: unable to get inventory: %w", op, err)
	}
	defer rows.Close()

	var inventory []InventoryItem
	for rows.Next() {
		var item InventoryItem
		if err := rows.Scan(&item.Type, &item.Quantity); err != nil {
			return InfoResponse{}, fmt.Errorf("%v: unable to get items: %w", op, err)
		}
		inventory = append(inventory, item)
	}

	var received, sent []Transaction
	receivedRows, err := s.db.QueryContext(
		ctx,
		"SELECT counterparty, amount FROM transactions WHERE user_id = $1 AND type = 'received'",
		userID,
	)
	if err != nil {
		return InfoResponse{}, fmt.Errorf("%v: unable to get received transactions: %w", op, err)
	}
	defer receivedRows.Close()

	for receivedRows.Next() {
		var t Transaction
		if err := receivedRows.Scan(&t.FromUser, &t.Amount); err != nil {
			return InfoResponse{}, fmt.Errorf("%v: unable to get details of received transactions: %w", op, err)
		}
		received = append(received, t)
	}

	sentRows, err := s.db.QueryContext(
		ctx,
		"SELECT counterparty, amount FROM transactions WHERE user_id = $1 AND type = 'sent'",
		userID,
	)
	if err != nil {
		return InfoResponse{}, fmt.Errorf("%v: unable to get sent transactions: %w", op, err)
	}
	defer sentRows.Close()

	for sentRows.Next() {
		var t Transaction
		if err := sentRows.Scan(&t.ToUser, &t.Amount); err != nil {
			return InfoResponse{}, fmt.Errorf("%v: unable to get details of sent transactions: %w", op, err)
		}
		sent = append(sent, t)
	}

	return InfoResponse{
		Coins:     coins,
		Inventory: inventory,
		CoinHistory: CoinHistory{
			Received: received,
			Sent:     sent,
		},
	}, nil
}
