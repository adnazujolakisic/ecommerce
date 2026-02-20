package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/metalbear-co/metalmart/services/order/models"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

func (s *PostgresStore) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS orders (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		order_number VARCHAR(50) UNIQUE,
		customer_email VARCHAR(255),
		customer_name VARCHAR(255),
		shipping_address JSONB,
		total_amount DECIMAL(10,2),
		status VARCHAR(50) DEFAULT 'pending',
		tracking_token UUID DEFAULT gen_random_uuid(),
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_orders_email ON orders(customer_email);
	CREATE INDEX IF NOT EXISTS idx_orders_token ON orders(tracking_token);
	CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);

	CREATE TABLE IF NOT EXISTS order_items (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
		product_id VARCHAR(50),
		product_name VARCHAR(255),
		quantity INTEGER,
		price_at_time DECIMAL(10,2),
		created_at TIMESTAMP DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
	`
	if _, err := s.db.Exec(query); err != nil {
		return err
	}
	// Add processor_source and source_topic for mirrord/Kafka queue splitting demo visibility
	if _, err := s.db.Exec(`ALTER TABLE orders ADD COLUMN IF NOT EXISTS processor_source VARCHAR(100)`); err != nil {
		return err
	}
	_, err := s.db.Exec(`ALTER TABLE orders ADD COLUMN IF NOT EXISTS source_topic VARCHAR(255)`)
	return err
}

func generateOrderNumber() string {
	return fmt.Sprintf("MM-%d-%s", time.Now().Unix(), uuid.New().String()[:8])
}

func (s *PostgresStore) CreateOrder(req models.CreateOrderRequest) (*models.Order, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var totalAmount float64
	for _, item := range req.Items {
		totalAmount += item.Price * float64(item.Quantity)
	}

	addressJSON, err := json.Marshal(req.ShippingAddress)
	if err != nil {
		return nil, err
	}

	orderNumber := generateOrderNumber()
	var order models.Order

	err = tx.QueryRow(`
		INSERT INTO orders (order_number, customer_email, customer_name, shipping_address, total_amount, status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
		RETURNING id, order_number, customer_email, customer_name, shipping_address, total_amount, status, tracking_token, created_at, updated_at
	`, orderNumber, req.CustomerEmail, req.CustomerName, addressJSON, totalAmount).Scan(
		&order.ID, &order.OrderNumber, &order.CustomerEmail, &order.CustomerName,
		&addressJSON, &order.TotalAmount, &order.Status, &order.TrackingToken,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(addressJSON, &order.ShippingAddress)

	for _, item := range req.Items {
		var orderItem models.OrderItem
		err = tx.QueryRow(`
			INSERT INTO order_items (order_id, product_id, product_name, quantity, price_at_time)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, order_id, product_id, product_name, quantity, price_at_time, created_at
		`, order.ID, item.ProductID, item.ProductName, item.Quantity, item.Price).Scan(
			&orderItem.ID, &orderItem.OrderID, &orderItem.ProductID, &orderItem.ProductName,
			&orderItem.Quantity, &orderItem.PriceAtTime, &orderItem.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		order.Items = append(order.Items, orderItem)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *PostgresStore) GetOrder(id string) (*models.Order, error) {
	order, err := s.getOrderByQuery(`SELECT id, order_number, customer_email, customer_name, shipping_address, total_amount, status, tracking_token, created_at, updated_at FROM orders WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	s.fillProcessorSource(order, true, id)
	return order, nil
}

func (s *PostgresStore) GetOrderByToken(token string) (*models.Order, error) {
	order, err := s.getOrderByQuery(`SELECT id, order_number, customer_email, customer_name, shipping_address, total_amount, status, tracking_token, created_at, updated_at FROM orders WHERE tracking_token = $1`, token)
	if err != nil {
		return nil, err
	}
	s.fillProcessorSource(order, false, token)
	return order, nil
}

func (s *PostgresStore) getOrderByQuery(query string, arg interface{}) (*models.Order, error) {
	var order models.Order
	var addressJSON []byte

	err := s.db.QueryRow(query, arg).Scan(
		&order.ID, &order.OrderNumber, &order.CustomerEmail, &order.CustomerName,
		&addressJSON, &order.TotalAmount, &order.Status, &order.TrackingToken,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(addressJSON, &order.ShippingAddress)

	items, err := s.getOrderItems(order.ID)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return &order, nil
}

func (s *PostgresStore) fillProcessorSource(order *models.Order, byID bool, val string) {
	col := "tracking_token"
	if byID {
		col = "id"
	}
	_ = s.db.QueryRow(`SELECT COALESCE(processor_source, ''), COALESCE(source_topic, '') FROM orders WHERE `+col+` = $1`, val).Scan(&order.ProcessedBy, &order.SourceTopic)
}

func (s *PostgresStore) ListOrders() ([]models.Order, error) {
	rows, err := s.db.Query(`
		SELECT id, order_number, customer_email, customer_name, shipping_address, total_amount, status, tracking_token, created_at, updated_at
		FROM orders ORDER BY created_at DESC LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanOrders(rows)
}

func (s *PostgresStore) ListOrdersByEmail(email string) ([]models.Order, error) {
	rows, err := s.db.Query(`
		SELECT id, order_number, customer_email, customer_name, shipping_address, total_amount, status, tracking_token, created_at, updated_at
		FROM orders WHERE customer_email = $1 ORDER BY created_at DESC
	`, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanOrders(rows)
}

func (s *PostgresStore) scanOrders(rows *sql.Rows) ([]models.Order, error) {
	var orders []models.Order
	for rows.Next() {
		var order models.Order
		var addressJSON []byte
		err := rows.Scan(
			&order.ID, &order.OrderNumber, &order.CustomerEmail, &order.CustomerName,
			&addressJSON, &order.TotalAmount, &order.Status, &order.TrackingToken,
			&order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(addressJSON, &order.ShippingAddress)
		_ = s.db.QueryRow(`SELECT COALESCE(processor_source, ''), COALESCE(source_topic, '') FROM orders WHERE id = $1`, order.ID).Scan(&order.ProcessedBy, &order.SourceTopic)
		orders = append(orders, order)
	}
	if orders == nil {
		orders = []models.Order{}
	}
	return orders, rows.Err()
}

func (s *PostgresStore) getOrderItems(orderID string) ([]models.OrderItem, error) {
	rows, err := s.db.Query(`
		SELECT id, order_id, product_id, product_name, quantity, price_at_time, created_at
		FROM order_items WHERE order_id = $1
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductName, &item.Quantity, &item.PriceAtTime, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []models.OrderItem{}
	}
	return items, rows.Err()
}

func (s *PostgresStore) GetOrderStatus(id string) (string, error) {
	var status string
	err := s.db.QueryRow(`SELECT status FROM orders WHERE id = $1`, id).Scan(&status)
	return status, err
}

func (s *PostgresStore) GetOrderStatusWithSource(id string) (status, processedBy, sourceTopic, customerEmail string, err error) {
	err = s.db.QueryRow(`
		SELECT status, COALESCE(customer_email, ''), COALESCE(processor_source, ''), COALESCE(source_topic, '')
		FROM orders WHERE id = $1
	`, id).Scan(&status, &customerEmail, &processedBy, &sourceTopic)
	if err != nil {
		return "", "", "", "", err
	}
	return status, processedBy, sourceTopic, customerEmail, nil
}

func (s *PostgresStore) UpdateOrderStatus(id, status, processorSource, sourceTopic string) error {
	if processorSource != "" {
		_, err := s.db.Exec(`UPDATE orders SET status = $1, processor_source = $2, source_topic = $3, updated_at = NOW() WHERE id = $4`, status, processorSource, sourceTopic, id)
		return err
	}
	_, err := s.db.Exec(`UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	return err
}
