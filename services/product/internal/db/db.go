package db

import "product/internal/entities"

type Db interface {
	CreateProduct(sku string, name string, description string, price float64, stock int64) (string, error)
	ListProducts(limit int, offset int) ([]entities.Product, error)
	DeleteProductById(id string) (bool, error)
	UpdateProductById(id string, detailsToUpdate map[string]interface{}) (bool, error)
	GetProductById(id string) (entities.Product, error)
}
