package db

import "order/internal/entities"

type Db interface {
	CreateOrder(userId string, status string, total float64) (string, error)
	CreateOrderItems(orderItems []entities.OrderItem) (bool, error)
	DeleteOrderById(id string) (bool, error)
	GetOrderById(id string) (entities.Order, error)
	GetOrderItemsByOrderId(orderId string) ([]entities.OrderItem, error)
	ListOrders(limit int, offset int) ([]entities.OrderDetails, error)
	DeleteOrderItemsByOrderId(orderId string) (bool, error)
	UpdateOrderById(id string, detailsToUpdate map[string]interface{}) (bool, error)

	//GetOrderDetailsWithItems(orderId string) (interface{}, error)  // uses DB join query
}
