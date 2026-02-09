package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/metalbear-co/metalmart/services/catalogue/handlers"
	"github.com/metalbear-co/metalmart/services/catalogue/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/catalogue?sslmode=disable"
	}

	db, err := store.NewPostgresStore(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if os.Getenv("SEED_DATA") == "true" {
		if err := db.Seed(); err != nil {
			log.Printf("Warning: Failed to seed data: %v", err)
		}
	}

	h := handlers.NewHandler(db)

	r := mux.NewRouter()

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/products", h.ListProducts).Methods("GET")
	api.HandleFunc("/products/search", h.SearchProducts).Methods("GET")
	api.HandleFunc("/products/category/{category}", h.ListByCategory).Methods("GET")
	api.HandleFunc("/products/{id}", h.GetProduct).Methods("GET")

	// CORS middleware
	handler := corsMiddleware(r)

	log.Printf("Catalogue service starting on port %s", port)
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
