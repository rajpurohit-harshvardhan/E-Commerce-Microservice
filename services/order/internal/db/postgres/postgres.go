package postgres

import (
	"common/utils/migrate"
	"database/sql"
	"embed"
	"errors"
	"log/slog"
	"order/internal/entities"
	"strings"

	"fmt"
	"order/internal/config"
	"time"

	_ "github.com/lib/pq"
)

type Postgres struct {
	Db *sql.DB
}

var embedMigrations embed.FS

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

	// Running migrations
	if err := migrate.Run(db, embedMigrations, "."); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	return &Postgres{
		Db: db,
	}, nil
}

func (p *Postgres) CreateOrder(userId string, status string, total float64) (string, error) {
	var id string

	err := p.Db.QueryRow(
		`INSERT INTO orders (user_id, status, total) VALUES ($1, $2, $3) RETURNING id`,
		userId, status, total).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

func (p *Postgres) CreateOrderItems(orderItems []entities.OrderItem) (bool, error) {
	if len(orderItems) == 0 {
		return false, nil // Nothing to insert
	}

	sqlQuery := `INSERT INTO order_items (order_id, product_id, quantity, price) VALUES `

	var queryValues []string
	var values []interface{}
	counter := 1
	for _, orderItem := range orderItems {
		queryValues = append(queryValues, fmt.Sprintf("($%d, $%d, $%d, $%d)", counter, counter+1, counter+2, counter+3))
		values = append(values, orderItem.OrderId, orderItem.ProductId, orderItem.Quantity, orderItem.Price)
		counter += 4
	}

	_, err := p.Db.Exec(sqlQuery+strings.Join(queryValues, ", "), values...)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (p *Postgres) DeleteOrderById(id string) (bool, error) {
	result, err := p.Db.Exec(`DELETE FROM orders where id=$1`, id)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		slog.Warn("failed to delete Order: ", err)
		return false, nil
	}

	return true, nil
}

func (p *Postgres) GetOrderById(id string) (entities.Order, error) {
	var order entities.Order
	err := p.Db.QueryRow("SELECT id, user_id, status, total, created_at, updated_at FROM orders WHERE id=$1",
		id).Scan(&order.ID, &order.UserId, &order.Status, &order.Total, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return order, err
	}
	return order, nil
}

func (p *Postgres) GetOrderItemsByOrderId(orderId string) ([]entities.OrderItem, error) {
	orderItem := []entities.OrderItem{}

	rows, err := p.Db.Query("SELECT id, order_id, product_id, quantity, price FROM order_items WHERE order_id=$1",
		orderId)
	if err != nil {
		return orderItem, err
	}

	return scanOrderItems(rows)
}

func scanOrderItems(rows *sql.Rows) ([]entities.OrderItem, error) {
	defer rows.Close()

	orderItems := []entities.OrderItem{}

	for rows.Next() {
		var item entities.OrderItem
		err := rows.Scan(
			&item.ID,
			&item.OrderId,
			&item.ProductId,
			&item.Quantity,
			&item.Price,
		)
		if err != nil {
			return nil, err
		}
		orderItems = append(orderItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orderItems, nil
}

func (p *Postgres) ListOrders(limit int, offset int) ([]entities.OrderDetails, error) {
	rows, err := p.Db.Query(`SELECT
       o.id, o.user_id, o.status, o.total, o.created_at, o.updated_at,
       oi.id as "itemId", oi.order_id, oi.product_id, oi.quantity, oi.price
   FROM
       orders o
   LEFT JOIN
       order_items oi ON o.id = oi.order_id
	LIMIT $1
	OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}

	return scanOrderDetails(rows)
}

func scanOrderDetails(rows *sql.Rows) ([]entities.OrderDetails, error) {
	defer rows.Close()

	groupedOrders := make(map[string]entities.OrderDetails)

	for rows.Next() {
		var currentOrder entities.Order
		var currentItem entities.OrderItem

		err := rows.Scan(
			&currentOrder.ID, &currentOrder.UserId, &currentOrder.Status, &currentOrder.Total, &currentOrder.CreatedAt, &currentOrder.UpdatedAt, // Order fields
			&currentItem.ID, &currentItem.OrderId, &currentItem.ProductId, &currentItem.Quantity, &currentItem.Price, // OrderItem fields
		)
		if err != nil {
			return nil, err
		}

		orderID := currentOrder.ID
		details, exists := groupedOrders[orderID]

		if !exists {
			details = entities.OrderDetails{
				Order: currentOrder,
				Items: make([]entities.OrderItem, 0),
			}
		}

		details.Items = append(details.Items, currentItem)
		groupedOrders[orderID] = details
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	orderDetailsList := make([]entities.OrderDetails, 0, len(groupedOrders))
	for _, details := range groupedOrders {
		orderDetailsList = append(orderDetailsList, details)
	}

	return orderDetailsList, nil
}

func (p *Postgres) DeleteOrderItemsByOrderId(orderId string) (bool, error) {
	result, err := p.Db.Exec(`DELETE FROM order_items where order_id=$1`, orderId)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		slog.Warn("failed to delete Order: ", err)
		return false, nil
	}

	return true, nil
}

func (p *Postgres) UpdateOrderById(id string, detailsToUpdate map[string]interface{}) (bool, error) {
	if len(detailsToUpdate) == 0 {
		return false, errors.New("no fields to update")
	}

	validColumns := map[string]bool{
		"status": true,
		"total":  true,
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

	query := fmt.Sprintf("UPDATE orders SET %s WHERE id = $%d", strings.Join(fields, ", "), placeholderCounter)
	values = append(values, id)

	result, err := p.Db.Exec(query, values...)
	if err != nil {
		return false, fmt.Errorf("failed to update auth: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		slog.Warn("failed to update auth: ", err)
		return false, fmt.Errorf("auth not found or no changes made")
	}

	return true, nil
}

/*func (p *Postgres) GetOrderDetailsWithItems(orderId string) (interface{}, error) {
	rows, err := p.Db.Query(`SELECT
       o.id, o.user_id, o.status, o.total, o.created_at, o.updated_at,
       oi.id as "itemId", oi.order_id, oi.product_id, oi.quantity, oi.price
   FROM
       orders o
   LEFT JOIN
       order_items oi ON o.id = oi.order_id
   WHERE
       o.id = $1`, orderId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var auth entities.Order
	var orderItems []entities.OrderItem
	var item entities.OrderItem

	orderFound := false

	for rows.Next() {
		err := rows.Scan(
			&auth.ID, &auth.UserId, &auth.Status, &auth.Total, &auth.CreatedAt, &auth.UpdatedAt, // Order fields
			&item.ID, &item.OrderId, &item.ProductId, &item.Quantity, &item.Price, // OrderItem fields
		)
		if err != nil {
			return nil, err
		}

		orderFound = true
		orderItems = append(orderItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if !orderFound {
		return nil, fmt.Errorf("auth not found")
	}

	finalResult := map[string]interface{}{
		"id":        auth.ID,
		"userId":    auth.UserId,
		"status":    auth.Status,
		"total":     auth.Total,
		"createdAt": auth.CreatedAt,
		"updatedAt": auth.UpdatedAt,
		"items":     orderItems,
	}

	return finalResult, nil
}*/
