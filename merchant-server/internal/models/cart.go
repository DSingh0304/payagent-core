package models

type CartItem struct {
	ProductID  string  `json:"product_id"`
	Name       string  `json:"name"`
	PricePaise int64   `json:"price_paise"`
	PriceINR   float64 `json:"price_inr"`
	Quantity   int     `json:"quantity"`
	Reasoning  string  `json:"reasoning"`
	ImageURL   string  `json:"image_url"`
}

type Cart struct {
	SessionID  string     `json:"session_id"`
	Items      []CartItem `json:"items"`
	TotalPaise int64      `json:"total_paise"`
	TotalINR   float64    `json:"total_inr"`
}
