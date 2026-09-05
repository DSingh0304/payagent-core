package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/DSingh0304/payagent-core/merchant-server/internal/models"
)

type CartService struct {
	Redis   *redis.Client
	Catalog *CatalogService
	mu      sync.Mutex
}

func NewCartService(rdb *redis.Client, catalog *CatalogService) *CartService {
	return &CartService{Redis: rdb, Catalog: catalog}
}

func (s *CartService) cartKey(sessionID string) string {
	return fmt.Sprintf("cart:%s", sessionID)
}

func (s *CartService) Get(ctx context.Context, sessionID string) (*models.Cart, error) {
	data, err := s.Redis.Get(ctx, s.cartKey(sessionID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return &models.Cart{SessionID: sessionID, Items: []models.CartItem{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var cart models.Cart
	json.Unmarshal(data, &cart)
	return &cart, nil
}

func (s *CartService) save(ctx context.Context, cart *models.Cart) error {
	data, _ := json.Marshal(cart)
	// Cache cart state in Redis with a TTL to prevent stale session accumulation
	return s.Redis.Set(ctx, s.cartKey(cart.SessionID), data, 2*time.Hour).Err()
}

func (s *CartService) AddItem(ctx context.Context, sessionID, productID string, qty int, reasoning string) (*models.Cart, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	product, err := s.Catalog.GetByID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	if product.Stock < qty {
		return nil, fmt.Errorf("STOCK_UNAVAILABLE")
	}

	cart, _ := s.Get(ctx, sessionID)

	found := false
	for i, item := range cart.Items {
		if item.ProductID == productID {
			cart.Items[i].Quantity += qty
			found = true
			break
		}
	}
	
	if !found {
		cart.Items = append(cart.Items, models.CartItem{
			ProductID:  product.ID,
			Name:       product.Name,
			PricePaise: product.PricePaise,
			PriceINR:   product.PriceINR,
			Quantity:   qty,
			Reasoning:  reasoning,
			ImageURL:   product.ImageURL,
		})
	}

	cart.TotalPaise = 0
	for _, item := range cart.Items {
		cart.TotalPaise += item.PricePaise * int64(item.Quantity)
	}
	cart.TotalINR = float64(cart.TotalPaise) / 100

	return cart, s.save(ctx, cart)
}

func (s *CartService) RemoveItem(ctx context.Context, sessionID, productID string, quantityToRemove int) (*models.Cart, error) {
	// Lock prevents race conditions during concurrent agent tool executions.
	s.mu.Lock()
	defer s.mu.Unlock()

	cart, _ := s.Get(ctx, sessionID)
	newItems := []models.CartItem{}
	
	for _, item := range cart.Items {
		if item.ProductID == productID {
			// If quantityToRemove is specified, we decrement. Otherwise, it acts as a full removal.
			if quantityToRemove > 0 {
				item.Quantity -= quantityToRemove
			} else {
				item.Quantity = 0
			}
			
			if item.Quantity > 0 {
				newItems = append(newItems, item)
			}
		} else {
			newItems = append(newItems, item)
		}
	}
	cart.Items = newItems
	
	cart.TotalPaise = 0
	for _, item := range cart.Items {
		cart.TotalPaise += item.PricePaise * int64(item.Quantity)
	}
	cart.TotalINR = float64(cart.TotalPaise) / 100
	
	return cart, s.save(ctx, cart)
}
