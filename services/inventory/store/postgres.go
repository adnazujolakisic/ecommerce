package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/metalbear-co/metalmart/services/inventory/models"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

func (s *PostgresStore) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS inventory (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		product_id VARCHAR(50) NOT NULL UNIQUE,
		stock_quantity INTEGER DEFAULT 0,
		reserved_quantity INTEGER DEFAULT 0,
		last_updated TIMESTAMP DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_inventory_product_id ON inventory(product_id);

	CREATE TABLE IF NOT EXISTS reservations (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		reservation_id UUID NOT NULL,
		product_id VARCHAR(50) NOT NULL,
		quantity INTEGER NOT NULL,
		status VARCHAR(20) DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_reservations_status ON reservations(status);
	CREATE INDEX IF NOT EXISTS idx_reservations_reservation_id ON reservations(reservation_id);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) SeedFromCatalogue(catalogueURL string) error {
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM inventory").Scan(&count)
	if count > 0 {
		log.Println("Inventory already seeded, skipping")
		return nil
	}

	resp, err := http.Get(catalogueURL + "/api/products")
	if err != nil {
		return fmt.Errorf("failed to fetch catalogue: %w", err)
	}
	defer resp.Body.Close()

	var products []struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		return fmt.Errorf("failed to decode products: %w", err)
	}

	for _, p := range products {
		s.InitInventory(p.ID, 100)
	}
	log.Printf("Seeded inventory with %d products", len(products))
	return nil
}

func (s *PostgresStore) GetInventory(productID string) (*models.Inventory, error) {
	var inv models.Inventory
	err := s.db.QueryRow(`
		SELECT id, product_id, stock_quantity, reserved_quantity, last_updated
		FROM inventory
		WHERE product_id = $1
	`, productID).Scan(&inv.ID, &inv.ProductID, &inv.StockQuantity, &inv.ReservedQuantity, &inv.LastUpdated)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (s *PostgresStore) InitInventory(productID string, quantity int) error {
	_, err := s.db.Exec(`
		INSERT INTO inventory (product_id, stock_quantity, reserved_quantity)
		VALUES ($1, $2, 0)
		ON CONFLICT (product_id) DO UPDATE SET stock_quantity = $2, last_updated = NOW()
	`, productID, quantity)
	return err
}

func (s *PostgresStore) Reserve(items []models.ReserveItem) (string, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	reservationID := uuid.New().String()

	for _, item := range items {
		var available int
		err := tx.QueryRow(`
			SELECT stock_quantity - reserved_quantity
			FROM inventory
			WHERE product_id = $1
			FOR UPDATE
		`, item.ProductID).Scan(&available)
		if err != nil {
			return "", fmt.Errorf("product %s not found in inventory", item.ProductID)
		}

		if available < item.Quantity {
			return "", fmt.Errorf("insufficient stock for product %s: available %d, requested %d", item.ProductID, available, item.Quantity)
		}

		_, err = tx.Exec(`
			UPDATE inventory
			SET reserved_quantity = reserved_quantity + $1, last_updated = NOW()
			WHERE product_id = $2
		`, item.Quantity, item.ProductID)
		if err != nil {
			return "", err
		}

		_, err = tx.Exec(`
			INSERT INTO reservations (reservation_id, product_id, quantity, status)
			VALUES ($1, $2, $3, 'pending')
		`, reservationID, item.ProductID, item.Quantity)
		if err != nil {
			return "", err
		}
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return reservationID, nil
}

func (s *PostgresStore) Release(reservationID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
		SELECT product_id, quantity FROM reservations
		WHERE reservation_id = $1 AND status = 'pending'
	`, reservationID)
	if err != nil {
		return err
	}

	var items []models.ReserveItem
	for rows.Next() {
		var item models.ReserveItem
		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			rows.Close()
			return err
		}
		items = append(items, item)
	}
	rows.Close()

	if len(items) == 0 {
		return fmt.Errorf("reservation not found or already processed")
	}

	for _, item := range items {
		_, err = tx.Exec(`
			UPDATE inventory
			SET reserved_quantity = reserved_quantity - $1, last_updated = NOW()
			WHERE product_id = $2
		`, item.Quantity, item.ProductID)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(`UPDATE reservations SET status = 'released' WHERE reservation_id = $1`, reservationID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *PostgresStore) Confirm(reservationID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
		SELECT product_id, quantity FROM reservations
		WHERE reservation_id = $1 AND status = 'pending'
	`, reservationID)
	if err != nil {
		return err
	}

	var items []models.ReserveItem
	for rows.Next() {
		var item models.ReserveItem
		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			rows.Close()
			return err
		}
		items = append(items, item)
	}
	rows.Close()

	if len(items) == 0 {
		return fmt.Errorf("reservation not found or already processed")
	}

	for _, item := range items {
		_, err = tx.Exec(`
			UPDATE inventory
			SET stock_quantity = stock_quantity - $1,
			    reserved_quantity = reserved_quantity - $1,
			    last_updated = NOW()
			WHERE product_id = $2
		`, item.Quantity, item.ProductID)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(`UPDATE reservations SET status = 'confirmed' WHERE reservation_id = $1`, reservationID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
