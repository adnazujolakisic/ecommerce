package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/metalbear-co/metalmart/services/order/kafka"
	"github.com/metalbear-co/metalmart/services/order/models"
	"github.com/metalbear-co/metalmart/services/order/store"
)

type Handler struct {
	store    *store.PostgresStore
	producer *kafka.Producer
}

func NewHandler(s *store.PostgresStore, p *kafka.Producer) *Handler {
	return &Handler{store: s, producer: p}
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	order, err := h.store.CreateOrder(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.producer != nil {
		event := models.OrderCreatedEvent{
			OrderID:       order.ID,
			OrderNumber:   order.OrderNumber,
			CustomerEmail: order.CustomerEmail,
			TotalAmount:   order.TotalAmount,
			Status:        order.Status,
			CreatedAt:     order.CreatedAt,
		}
		if err := h.producer.PublishOrderCreated(event); err != nil {
			// Log but don't fail the request
			println("Warning: Failed to publish order created event:", err.Error())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	order, err := h.store.GetOrder(id)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *Handler) GetOrderByToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]

	order, err := h.store.GetOrderByToken(token)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	var orders []models.Order
	var err error

	if email != "" {
		orders, err = h.store.ListOrdersByEmail(email)
	} else {
		orders, err = h.store.ListOrders()
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *Handler) GetOrderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	status, err := h.store.GetOrderStatus(id)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": status})
}

func (h *Handler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req models.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.store.UpdateOrderStatus(id, req.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": req.Status})
}
