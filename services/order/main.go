package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/metalbear-co/metalmart/services/order/handlers"
	"github.com/metalbear-co/metalmart/services/order/kafka"
	"github.com/metalbear-co/metalmart/services/order/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/orders?sslmode=disable"
	}

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	db, err := store.NewPostgresStore(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	producer, err := kafka.NewProducer(kafkaBrokers)
	if err != nil {
		log.Printf("Warning: Failed to connect to Kafka: %v", err)
		producer = nil
	}
	if producer != nil {
		defer producer.Close()
	}

	h := handlers.NewHandler(db, producer)

	r := mux.NewRouter()

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/orders", h.CreateOrder).Methods("POST")
	api.HandleFunc("/orders", h.ListOrders).Methods("GET")
	api.HandleFunc("/orders/{id}", h.GetOrder).Methods("GET")
	api.HandleFunc("/orders/{id}/status", h.GetOrderStatus).Methods("GET")
	api.HandleFunc("/orders/{id}/status", h.UpdateOrderStatus).Methods("PUT")
	api.HandleFunc("/orders/track/{token}", h.GetOrderByToken).Methods("GET")

	handler := corsMiddleware(r)

	log.Printf("Order service starting on port %s", port)
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
