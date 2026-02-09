package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/metalbear-co/metalmart/services/checkout/models"
)

type Handler struct {
	inventoryURL string
	orderURL     string
	httpClient   *http.Client
}

func NewHandler(inventoryURL, orderURL string) *Handler {
	return &Handler{
		inventoryURL: inventoryURL,
		orderURL:     orderURL,
		httpClient:   &http.Client{},
	}
}

func (h *Handler) ProcessCheckout(w http.ResponseWriter, r *http.Request) {
	var req models.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.CustomerEmail == "" || req.CustomerName == "" || len(req.Items) == 0 {
		respondError(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Step 1: Reserve inventory
	reserveReq := models.ReserveRequest{Items: make([]models.ReserveItem, len(req.Items))}
	for i, item := range req.Items {
		reserveReq.Items[i] = models.ReserveItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	reserveResp, err := h.reserveInventory(reserveReq)
	if err != nil {
		respondError(w, fmt.Sprintf("Failed to reserve inventory: %v", err), http.StatusConflict)
		return
	}

	if !reserveResp.Success {
		respondError(w, reserveResp.Message, http.StatusConflict)
		return
	}

	// Step 2: Create order
	orderReq := models.CreateOrderRequest{
		CustomerEmail:   req.CustomerEmail,
		CustomerName:    req.CustomerName,
		ShippingAddress: req.ShippingAddress,
		ReservationID:   reserveResp.ReservationID,
	}
	for _, item := range req.Items {
		orderReq.Items = append(orderReq.Items, models.OrderItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
		})
	}

	orderResp, err := h.createOrder(orderReq)
	if err != nil {
		h.releaseInventory(reserveResp.ReservationID)
		respondError(w, fmt.Sprintf("Failed to create order: %v", err), http.StatusInternalServerError)
		return
	}

	// Step 3: Confirm inventory reservation
	if err := h.confirmInventory(reserveResp.ReservationID); err != nil {
		// Log but don't fail - order is already created
		fmt.Printf("Warning: Failed to confirm inventory: %v\n", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.CheckoutResponse{
		Success:       true,
		OrderID:       orderResp.ID,
		OrderNumber:   orderResp.OrderNumber,
		TrackingToken: orderResp.TrackingToken,
		TotalAmount:   orderResp.TotalAmount,
	})
}

func (h *Handler) ValidateCart(w http.ResponseWriter, r *http.Request) {
	var req models.ValidateCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check inventory for each item
	for _, item := range req.Items {
		resp, err := h.httpClient.Get(fmt.Sprintf("%s/api/inventory/%s", h.inventoryURL, item.ProductID))
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(models.ValidateCartResponse{
				Valid:   false,
				Message: fmt.Sprintf("Failed to check inventory for %s", item.ProductName),
			})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(models.ValidateCartResponse{
				Valid:   false,
				Message: fmt.Sprintf("Product %s not found", item.ProductName),
			})
			return
		}

		var inv struct {
			StockQuantity    int `json:"stock_quantity"`
			ReservedQuantity int `json:"reserved_quantity"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&inv); err != nil {
			continue
		}

		available := inv.StockQuantity - inv.ReservedQuantity
		if available < item.Quantity {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(models.ValidateCartResponse{
				Valid:   false,
				Message: fmt.Sprintf("Not enough stock for %s (available: %d)", item.ProductName, available),
			})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.ValidateCartResponse{Valid: true})
}

func (h *Handler) reserveInventory(req models.ReserveRequest) (*models.ReserveResponse, error) {
	body, _ := json.Marshal(req)
	resp, err := h.httpClient.Post(
		fmt.Sprintf("%s/api/inventory/reserve", h.inventoryURL),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var reserveResp models.ReserveResponse
	if err := json.NewDecoder(resp.Body).Decode(&reserveResp); err != nil {
		return nil, err
	}

	return &reserveResp, nil
}

func (h *Handler) releaseInventory(reservationID string) error {
	body, _ := json.Marshal(map[string]string{"reservation_id": reservationID})
	resp, err := h.httpClient.Post(
		fmt.Sprintf("%s/api/inventory/release", h.inventoryURL),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (h *Handler) confirmInventory(reservationID string) error {
	body, _ := json.Marshal(map[string]string{"reservation_id": reservationID})
	resp, err := h.httpClient.Post(
		fmt.Sprintf("%s/api/inventory/confirm", h.inventoryURL),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (h *Handler) createOrder(req models.CreateOrderRequest) (*models.CreateOrderResponse, error) {
	body, _ := json.Marshal(req)
	resp, err := h.httpClient.Post(
		fmt.Sprintf("%s/api/orders", h.orderURL),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("order service error: %s", string(bodyBytes))
	}

	var orderResp models.CreateOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return nil, err
	}

	return &orderResp, nil
}

func respondError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.CheckoutResponse{
		Success: false,
		Message: message,
	})
}
