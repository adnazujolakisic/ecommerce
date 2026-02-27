package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/metalbear-co/metalmart/services/inventory/models"
	"github.com/metalbear-co/metalmart/services/inventory/store"
)

type Handler struct {
	store      *store.PostgresStore
	dbSource   string
}

func NewHandler(s *store.PostgresStore, dbSource string) *Handler {
	if dbSource == "" {
		if os.Getenv("MIRRORD_DB_BRANCH") == "true" {
			dbSource = "mirrord-db-branch"
		} else {
			dbSource = "cluster"
		}
	}
	return &Handler{store: s, dbSource: dbSource}
}

func (h *Handler) setDatabaseSourceHeader(w http.ResponseWriter) {
	w.Header().Set("X-Database-Source", h.dbSource)
}

func (h *Handler) GetInventory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["productId"]

	h.setDatabaseSourceHeader(w)

	inv, err := h.store.GetInventory(productID)
	if err != nil {
		http.Error(w, "Inventory not found", http.StatusNotFound)
		return
	}

	// Log only when INVENTORY_DEBUG=1 (avoids noisy GET spam from product listings)
	if os.Getenv("INVENTORY_DEBUG") == "1" {
		log.Printf("[%s] GetInventory product_id=%s → stock=%d reserved=%d", h.dbSource, productID, inv.StockQuantity, inv.ReservedQuantity)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inv)
}

func (h *Handler) Reserve(w http.ResponseWriter, r *http.Request) {
	h.setDatabaseSourceHeader(w)

	var req models.ReserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	reservationID, err := h.store.Reserve(req.Items)
	if err != nil {
		log.Printf("[%s] Reserve FAILED items=%v: %v", h.dbSource, req.Items, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(models.ReserveResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	log.Printf("[%s] Reserve OK items=%v → reservation_id=%s", h.dbSource, req.Items, reservationID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.ReserveResponse{
		ReservationID: reservationID,
		Success:       true,
	})
}

func (h *Handler) Release(w http.ResponseWriter, r *http.Request) {
	h.setDatabaseSourceHeader(w)
	var req models.ReleaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.store.Release(req.ReservationID); err != nil {
		log.Printf("[%s] Release FAILED reservation_id=%s: %v", h.dbSource, req.ReservationID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[%s] Release OK reservation_id=%s (reverted reserved qty)", h.dbSource, req.ReservationID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *Handler) Confirm(w http.ResponseWriter, r *http.Request) {
	h.setDatabaseSourceHeader(w)
	var req models.ConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.store.Confirm(req.ReservationID); err != nil {
		log.Printf("[%s] Confirm FAILED reservation_id=%s: %v", h.dbSource, req.ReservationID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[%s] Confirm OK reservation_id=%s (stock reduced on branch)", h.dbSource, req.ReservationID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *Handler) InitInventory(w http.ResponseWriter, r *http.Request) {
	h.setDatabaseSourceHeader(w)
	var req models.InitInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.store.InitInventory(req.ProductID, req.Quantity); err != nil {
		log.Printf("[%s] InitInventory FAILED product_id=%s qty=%d: %v", h.dbSource, req.ProductID, req.Quantity, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[%s] InitInventory OK product_id=%s qty=%d", h.dbSource, req.ProductID, req.Quantity)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
