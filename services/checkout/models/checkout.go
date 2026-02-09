package models

type CheckoutRequest struct {
	CustomerEmail   string          `json:"customer_email"`
	CustomerName    string          `json:"customer_name"`
	ShippingAddress ShippingAddress `json:"shipping_address"`
	Items           []CartItem      `json:"items"`
}

type ShippingAddress struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
	Country string `json:"country"`
}

type CartItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type CheckoutResponse struct {
	Success       bool   `json:"success"`
	OrderID       string `json:"order_id,omitempty"`
	OrderNumber   string `json:"order_number,omitempty"`
	TrackingToken string `json:"tracking_token,omitempty"`
	Message       string `json:"message,omitempty"`
	TotalAmount   float64 `json:"total_amount,omitempty"`
}

type ValidateCartRequest struct {
	Items []CartItem `json:"items"`
}

type ValidateCartResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
}

// Internal types for service communication
type ReserveRequest struct {
	Items []ReserveItem `json:"items"`
}

type ReserveItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type ReserveResponse struct {
	ReservationID string `json:"reservation_id"`
	Success       bool   `json:"success"`
	Message       string `json:"message,omitempty"`
}

type CreateOrderRequest struct {
	CustomerEmail   string          `json:"customer_email"`
	CustomerName    string          `json:"customer_name"`
	ShippingAddress ShippingAddress `json:"shipping_address"`
	Items           []OrderItem     `json:"items"`
	ReservationID   string          `json:"reservation_id"`
}

type OrderItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type CreateOrderResponse struct {
	ID            string  `json:"id"`
	OrderNumber   string  `json:"order_number"`
	TrackingToken string  `json:"tracking_token"`
	TotalAmount   float64 `json:"total_amount"`
	Status        string  `json:"status"`
}
