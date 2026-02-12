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
	// Debug: show which DB host we're connecting to (mirrord branch vs cluster)
	if host := extractDBHost(dbURL); host != "" {
		log.Printf("DB connection host: %s", host)
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

	dbSource := "cluster"
	if os.Getenv("MIRRORD_DB_BRANCH") == "true" {
		dbSource = "mirrord-db-branch"
	}
	h := handlers.NewHandler(db, dbSource)

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

	if dbSource == "mirrord-db-branch" {
		log.Printf("📦 Database: MIRRORD DB BRANCH (isolated copy - your changes don't affect cluster)")
	} else {
		log.Printf("📦 Database: CLUSTER (shared production database)")
	}
	log.Printf("Inventory service starting on port %s [%s]", port, dbSource)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func extractDBHost(url string) string {
	if i := strings.Index(url, "@"); i >= 0 && i+1 < len(url) {
		rest := url[i+1:]
		if j := strings.IndexAny(rest, ":/"); j >= 0 {
			return rest[:j]
		}
		return rest
	}
	return ""
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
