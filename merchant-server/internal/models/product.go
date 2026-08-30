package models

import "time"

type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	PricePaise  int64     `json:"price_paise"`
	PriceINR    float64   `json:"price_inr"`
	Stock       int       `json:"stock"`
	Tags        []string  `json:"tags"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
}
