package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	"github.com/metalbear-co/metalmart/services/order/kafka"
	"github.com/metalbear-co/metalmart/services/order/models"
	"github.com/metalbear-co/metalmart/services/order/store"
)

// queueSplittingEmailFilter must match .mirrord/queue-splitting.json customer_email filter.
// Only orders with matching emails are routed to mirrord's temp topic; only those should show the mirrord badge.
var queueSplittingEmailFilter = regexp.MustCompile(`.*@metalbear\.com`)

// setOrderSource computes and sets Source ("mirrord" or "cluster") on the order.
// Mirrord badge only when: (a) processed by mirrord session AND (b) email matches queue-splitting filter.
func setOrderSource(order *models.Order) {
	isMirrordProcessed := order.ProcessedBy == "mirrord-kafka" || strings.Contains(order.SourceTopic, "mirrord-tmp")
	emailMatchesFilter := order.CustomerEmail != "" && queueSplittingEmailFilter.MatchString(order.CustomerEmail)
	if isMirrordProcessed && emailMatchesFilter {
		order.Source = "mirrord"
	} else {
		order.Source = "cluster"
	}
}

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

	if b, _ := json.MarshalIndent(order, "", "  "); len(b) > 0 {
		log.Printf("Order created and inserted: %s", string(b))
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
			log.Printf("Warning: Failed to publish order created event: %v", err)
		} else if b, _ := json.MarshalIndent(event, "", "  "); len(b) > 0 {
			log.Printf("Published to Kafka: %s", string(b))
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

	setOrderSource(order)
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

	setOrderSource(order)
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

	for i := range orders {
		setOrderSource(&orders[i])
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *Handler) GetOrderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	status, processedBy, sourceTopic, customerEmail, err := h.store.GetOrderStatusWithSource(id)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	resp := map[string]string{"status": status}
	isMirrordProcessed := processedBy == "mirrord-kafka" || strings.Contains(sourceTopic, "mirrord-tmp")
	emailMatchesFilter := customerEmail != "" && queueSplittingEmailFilter.MatchString(customerEmail)
	if isMirrordProcessed && emailMatchesFilter {
		resp["processed_by"] = processedBy
		resp["source_topic"] = sourceTopic
		resp["source"] = "mirrord"
	} else {
		resp["source"] = "cluster"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req models.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	processorSource := r.Header.Get("X-Processor-Source")
	sourceTopic := r.Header.Get("X-Kafka-Topic")

	if err := h.store.UpdateOrderStatus(id, req.Status, processorSource, sourceTopic); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]string{"status": req.Status}
	if processorSource != "" {
		resp["processed_by"] = processorSource
		resp["source_topic"] = sourceTopic
		resp["source"] = "mirrord"
		w.Header().Set("X-Processed-By", processorSource)
		w.Header().Set("X-Source-Topic", sourceTopic)
	} else {
		resp["source"] = "cluster"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
