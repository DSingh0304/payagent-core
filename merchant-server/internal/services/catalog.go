package services

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DSingh0304/payagent-core/merchant-server/internal/models"
)

type CatalogService struct {
	DB *pgxpool.Pool
}

func NewCatalogService(db *pgxpool.Pool) *CatalogService {
	return &CatalogService{DB: db}
}

func (s *CatalogService) Search(ctx context.Context, query, category string, maxPricePaise int64) ([]models.Product, error) {
	sql := `SELECT id, name, description, category, price_paise, stock, tags, image_url, created_at FROM products WHERE 1=1`
	args := []interface{}{}
	i := 1

	if query != "" {
		sql += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", i, i)
		args = append(args, "%"+query+"%")
		i++
	}
	if category != "" {
		sql += fmt.Sprintf(" AND category = $%d", i)
		args = append(args, category)
		i++
	}
	if maxPricePaise > 0 {
		sql += fmt.Sprintf(" AND price_paise <= $%d", i)
		args = append(args, maxPricePaise)
		i++
	}
	sql += " ORDER BY price_paise ASC"

	rows, err := s.DB.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		rows.Scan(&p.ID, &p.Name, &p.Description, &p.Category, &p.PricePaise, &p.Stock, &p.Tags, &p.ImageURL, &p.CreatedAt)
		p.PriceINR = float64(p.PricePaise) / 100
		products = append(products, p)
	}
	return products, nil
}

func (s *CatalogService) GetByID(ctx context.Context, id string) (*models.Product, error) {
	var p models.Product
	err := s.DB.QueryRow(ctx, `
		SELECT id, name, description, category, price_paise, stock, tags, image_url, created_at
		FROM products WHERE id = $1
	`, id).Scan(&p.ID, &p.Name, &p.Description, &p.Category, &p.PricePaise, &p.Stock, &p.Tags, &p.ImageURL, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	p.PriceINR = float64(p.PricePaise) / 100
	return &p, nil
}
