package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/metalbear-co/metalmart/services/checkout/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	inventoryURL := os.Getenv("INVENTORY_SERVICE_URL")
	if inventoryURL == "" {
		inventoryURL = "http://localhost:8082"
	}

	orderURL := os.Getenv("ORDER_SERVICE_URL")
	if orderURL == "" {
		orderURL = "http://localhost:8084"
	}

	h := handlers.NewHandler(inventoryURL, orderURL)

	r := mux.NewRouter()

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/checkout", h.ProcessCheckout).Methods("POST")
	api.HandleFunc("/checkout/validate", h.ValidateCart).Methods("POST")

	handler := corsMiddleware(r)

	log.Printf("Checkout service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
