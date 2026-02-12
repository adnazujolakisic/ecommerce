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

	log.Printf("[%s] GetInventory product_id=%s", h.dbSource, productID)
	h.setDatabaseSourceHeader(w)

	inv, err := h.store.GetInventory(productID)
	if err != nil {
		http.Error(w, "Inventory not found", http.StatusNotFound)
		return
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

	log.Printf("[%s] Reserve items=%v", h.dbSource, req.Items)

	reservationID, err := h.store.Reserve(req.Items)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(models.ReserveResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.ReserveResponse{
		ReservationID: reservationID,
		Success:       true,
	})
}

func (h *Handler) Release(w http.ResponseWriter, r *http.Request) {
	var req models.ReleaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.store.Release(req.ReservationID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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

	log.Printf("[%s] Confirm reservation_id=%s", h.dbSource, req.ReservationID)
	if err := h.store.Confirm(req.ReservationID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *Handler) InitInventory(w http.ResponseWriter, r *http.Request) {
	var req models.InitInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.store.InitInventory(req.ProductID, req.Quantity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
