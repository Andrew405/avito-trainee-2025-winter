package models

import (
	"time"
)

type Transaction struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	Type         string    `json:"type"`                   // "sent", "received", "purchased"
	Counterparty string    `json:"counterparty,omitempty"` // Имя контрагента (если применимо)
	Merch        string    `json:"merch,omitempty"`        // Название товара (если применимо)
	Amount       int       `json:"amount"`
	CreatedAt    time.Time `json:"created_at"`
}
