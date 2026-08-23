package models

import "time"

type UPIMandate struct {
	ID             string    `json:"id"`
	SessionID      string    `json:"session_id"`
	CustomerID     string    `json:"customer_id"`
	TokenID        string    `json:"token_id"`
	BlockOrderID   string    `json:"block_order_id"`
	MaxAmountPaise int64     `json:"max_amount_paise"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}
