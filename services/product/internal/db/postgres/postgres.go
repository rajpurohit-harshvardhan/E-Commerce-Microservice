package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"product/internal/config"
	"product/internal/entities"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type Postgres struct {
	Db *sql.DB
}

func New(cfg *config.Config) (*Postgres, error) {
	// Build connection string
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Cockroach.User,
		cfg.Cockroach.Password,
		cfg.Cockroach.Host,
		cfg.Cockroach.Port,
		cfg.Cockroach.DBName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Configure the connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// table creation
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS products ( 
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     sku STRING UNIQUE NOT NULL,
     name STRING NOT NULL,
     description STRING,
     price DECIMAL(10,2) NOT NULL,
     stock INT8 DEFAULT 0,
     created_at TIMESTAMPTZ DEFAULT now(),
     updated_at TIMESTAMPTZ DEFAULT now());`)
	if err != nil {
		return nil, err
	}

	return &Postgres{
		Db: db,
	}, nil
}

func (p *Postgres) CreateProduct(sku string, name string, description string, price float64, stock int64) (string, error) {
	var id string

	err := p.Db.QueryRow(
		`INSERT INTO products (sku, name, description, price, stock)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		sku, name, description, price, stock,
	).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

func (p *Postgres) ListProducts(limit int, offset int) ([]entities.Product, error) {
	rows, err := p.Db.Query("SELECT id, sku, name, description, price, stock, created_at, updated_at"+
		" FROM products LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, err
	}
	return scanProducts(rows)
}

func scanProducts(rows *sql.Rows) ([]entities.Product, error) {
	defer rows.Close() // Ensure rows are closed

	products := []entities.Product{}

	for rows.Next() {
		var product entities.Product
		err := rows.Scan(
			&product.ID,
			&product.SKU,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Stock,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	// Check for any errors that occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (p *Postgres) DeleteProductById(id string) (bool, error) {
	_, err := p.Db.Query(`DELETE FROM products where id=$1`, id)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (p *Postgres) UpdateProductById(id string, detailsToUpdate map[string]interface{}) (bool, error) {
	if len(detailsToUpdate) == 0 {
		return false, errors.New("no fields to update")
	}

	validColumns := map[string]bool{
		"sku":         true,
		"name":        true,
		"description": true,
		"price":       true,
		"stock":       true,
	}

	var fields []string
	var values []interface{}
	placeholderCounter := 1

	for key, value := range detailsToUpdate {
		// Validate the field name against the valid columns
		if !validColumns[key] {
			return false, fmt.Errorf("invalid field: %s", key)
		}
		fields = append(fields, fmt.Sprintf("%s = $%d", key, placeholderCounter))
		values = append(values, value)
		placeholderCounter++
	}

	// Append `updated_at` for good measure, and handle its placeholder.
	fields = append(fields, fmt.Sprintf("updated_at = $%d", placeholderCounter))
	values = append(values, time.Now())
	placeholderCounter++

	query := fmt.Sprintf("UPDATE products SET %s WHERE id = $%d", strings.Join(fields, ", "), placeholderCounter)
	values = append(values, id)

	result, err := p.Db.Exec(query, values...)
	if err != nil {
		return false, fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		slog.Warn("failed to update product: ", err)
		return false, fmt.Errorf("product not found or no changes made")
	}

	return true, nil
}

func (p *Postgres) GetProductById(id string) (entities.Product, error) {
	var product entities.Product
	err := p.Db.QueryRow("SELECT id, sku, name, description, price, stock, created_at, updated_at FROM products"+
		" WHERE id=$1", id).Scan(&product.ID, &product.SKU, &product.Name, &product.Description, &product.Price,
		&product.Stock, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		return product, err
	}
	return product, nil
}
