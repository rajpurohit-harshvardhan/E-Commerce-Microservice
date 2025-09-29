package entities

import "time"

type Product struct {
	ID          string    `json:"id"`
	SKU         string    `json:"sku" validate:"required"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description,omitempty"`
	Price       float64   `json:"price" validate:"required"`
	Stock       int64     `json:"stock" validate:"required"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
