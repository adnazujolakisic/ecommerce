package models

import "time"

type Inventory struct {
	ID               string    `json:"id"`
	ProductID        string    `json:"product_id"`
	StockQuantity    int       `json:"stock_quantity"`
	ReservedQuantity int       `json:"reserved_quantity"`
	LastUpdated      time.Time `json:"last_updated"`
}

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

type ReleaseRequest struct {
	ReservationID string `json:"reservation_id"`
}

type ConfirmRequest struct {
	ReservationID string `json:"reservation_id"`
}

type InitInventoryRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}
