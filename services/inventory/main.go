package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/metalbear-co/metalmart/services/inventory/handlers"
	"github.com/metalbear-co/metalmart/services/inventory/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/inventory?sslmode=disable"
	}
	// Ensure sslmode=disable for db branching compatibility
	if !strings.Contains(dbURL, "sslmode=") {
		if strings.Contains(dbURL, "?") {
			dbURL += "&sslmode=disable"
		} else {
			dbURL += "?sslmode=disable"
		}
	}

	db, err := store.NewPostgresStore(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	catalogueURL := os.Getenv("CATALOGUE_SERVICE_URL")
	if catalogueURL == "" {
		catalogueURL = "http://catalogue:8081"
	}
	if err := db.SeedFromCatalogue(catalogueURL); err != nil {
		log.Printf("Warning: Failed to seed inventory from catalogue: %v", err)
	}

	h := handlers.NewHandler(db)

	r := mux.NewRouter()

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/inventory/{productId}", h.GetInventory).Methods("GET")
	api.HandleFunc("/inventory/reserve", h.Reserve).Methods("POST")
	api.HandleFunc("/inventory/release", h.Release).Methods("POST")
	api.HandleFunc("/inventory/confirm", h.Confirm).Methods("POST")
	api.HandleFunc("/inventory/init", h.InitInventory).Methods("POST")

	handler := corsMiddleware(r)

	log.Printf("Inventory service starting on port %s", port)
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
