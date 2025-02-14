package models

import (
	"time"
)

type InventoryItem struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ItemType  string    `json:"item_type"`
	Quantity  int       `json:"quantity"`
	UpdatedAt time.Time `json:"updated_at"`
}
