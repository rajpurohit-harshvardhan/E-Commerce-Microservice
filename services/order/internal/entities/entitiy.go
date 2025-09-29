package entities

import "time"

type Order struct {
	ID        string    `json:"id"`
	UserId    string    `json:"user_id" validate:"required"`
	Status    string    `json:"status" validate:"required"` // PENDING/CONFIRMED/CANCELLED
	Total     float64   `json:"total" validate:"required"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrderItem struct {
	ID        string  `json:"id"`
	OrderId   string  `json:"order_id" validate:"required"`
	ProductId string  `json:"product_id" validate:"required"`
	Quantity  int64   `json:"quantity" validate:"required"`
	Price     float64 `json:"price" validate:"required"`
}

type OrderDetails struct {
	Order
	Items []OrderItem `json:"items" validate:"required"`
}

type CreateOrderRequestInputItems struct {
	ID       string  `json:"id"  validate:"required"`
	Quantity int64   `json:"quantity" validate:"required"`
	Price    float64 `json:"price" validate:"required"`
}

type CreateOrderRequestInput struct {
	Items []CreateOrderRequestInputItems `json:"items" validate:"required"`
}
