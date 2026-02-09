package models

import "time"

type Order struct {
	ID              string         `json:"id"`
	OrderNumber     string         `json:"order_number"`
	CustomerEmail   string         `json:"customer_email"`
	CustomerName    string         `json:"customer_name"`
	ShippingAddress ShippingAddress `json:"shipping_address"`
	TotalAmount     float64        `json:"total_amount"`
	Status          string         `json:"status"`
	TrackingToken   string         `json:"tracking_token"`
	Items           []OrderItem    `json:"items,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type ShippingAddress struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
	Country string `json:"country"`
}

type OrderItem struct {
	ID          string    `json:"id"`
	OrderID     string    `json:"order_id"`
	ProductID   string    `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	PriceAtTime float64   `json:"price_at_time"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateOrderRequest struct {
	CustomerEmail   string          `json:"customer_email"`
	CustomerName    string          `json:"customer_name"`
	ShippingAddress ShippingAddress `json:"shipping_address"`
	Items           []OrderItemInput `json:"items"`
	ReservationID   string          `json:"reservation_id"`
}

type OrderItemInput struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

type OrderCreatedEvent struct {
	OrderID       string    `json:"order_id"`
	OrderNumber   string    `json:"order_number"`
	CustomerEmail string    `json:"customer_email"`
	TotalAmount   float64   `json:"total_amount"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}
