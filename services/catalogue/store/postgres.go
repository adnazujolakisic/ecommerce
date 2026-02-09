package store

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/metalbear-co/metalmart/services/catalogue/models"
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
	CREATE TABLE IF NOT EXISTS products (
		id VARCHAR(50) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		price DECIMAL(10,2) NOT NULL,
		image_url VARCHAR(500),
		category VARCHAR(100),
		created_at TIMESTAMP DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
	CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) Seed() error {
	// Check if data already exists
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // Already seeded
	}

	products := []struct {
		id          string
		name        string
		description string
		price       float64
		imageURL    string
		category    string
	}{
		{
			id:          "1",
			name:        "MetalBear Classic T-Shirt",
			description: "Classic black t-shirt featuring the iconic MetalBear logo. Made from 100% organic cotton for maximum comfort during those long debugging sessions.",
			price:       29.99,
			imageURL:    "/images/tshirt-classic.svg",
			category:    "t-shirts",
		},
		{
			id:          "2",
			name:        "MetalBear Logo Hoodie",
			description: "Stay warm while coding with this premium hoodie. Features the MetalBear logo on the front and 'Debug Locally, Ship Globally' on the back.",
			price:       59.99,
			imageURL:    "/images/hoodie-logo.svg",
			category:    "hoodies",
		},
		{
			id:          "3",
			name:        "MetalBear Dev T-Shirt - Works on My Machine",
			description: "The classic developer excuse, now on a shirt. Perfect for standup meetings and code reviews.",
			price:       29.99,
			imageURL:    "/images/tshirt-works.svg",
			category:    "t-shirts",
		},
		{
			id:          "4",
			name:        "MetalBear Cap",
			description: "Adjustable snapback cap with embroidered MetalBear logo. Shield your eyes from the glare of your monitor.",
			price:       24.99,
			imageURL:    "/images/cap.svg",
			category:    "accessories",
		},
		{
			id:          "5",
			name:        "MetalBear Mug - Debug with Coffee",
			description: "Large 15oz ceramic mug perfect for your morning coffee or late-night debugging fuel. Microwave and dishwasher safe.",
			price:       14.99,
			imageURL:    "/images/mug.svg",
			category:    "accessories",
		},
		{
			id:          "6",
			name:        "MetalBear Sticker Pack",
			description: "Set of 10 high-quality vinyl stickers featuring MetalBear logos and developer humor. Perfect for laptops, water bottles, and keyboards.",
			price:       9.99,
			imageURL:    "/images/stickers.svg",
			category:    "accessories",
		},
		{
			id:          "7",
			name:        "MetalBear Polo Shirt",
			description: "Professional polo shirt with subtle MetalBear branding. Perfect for when you need to look presentable but still rep your favorite tools.",
			price:       44.99,
			imageURL:    "/images/polo.svg",
			category:    "t-shirts",
		},
		{
			id:          "8",
			name:        "MetalBear Zip Hoodie",
			description: "Premium zip-up hoodie with MetalBear logo. Features deep pockets for your phone and snacks. Available in charcoal grey.",
			price:       69.99,
			imageURL:    "/images/hoodie-zip.svg",
			category:    "hoodies",
		},
		{
			id:          "9",
			name:        "MetalBear Socks - 3 Pack",
			description: "Comfortable crew socks with MetalBear patterns. Because even your feet deserve good developer swag.",
			price:       19.99,
			imageURL:    "/images/socks.svg",
			category:    "accessories",
		},
		{
			id:          "10",
			name:        "MetalBear Beanie",
			description: "Warm knit beanie with embroidered MetalBear logo. Perfect for cold server rooms and winter debugging.",
			price:       22.99,
			imageURL:    "/images/beanie.svg",
			category:    "accessories",
		},
	}

	for _, p := range products {
		_, err := s.db.Exec(
			`INSERT INTO products (id, name, description, price, image_url, category) VALUES ($1, $2, $3, $4, $5, $6)`,
			p.id, p.name, p.description, p.price, p.imageURL, p.category,
		)
		if err != nil {
			return fmt.Errorf("failed to seed product %s: %w", p.name, err)
		}
	}

	return nil
}

func (s *PostgresStore) ListProducts() ([]models.Product, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, price, image_url, category, created_at
		FROM products
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanProducts(rows)
}

func (s *PostgresStore) GetProduct(id string) (*models.Product, error) {
	var p models.Product
	err := s.db.QueryRow(`
		SELECT id, name, description, price, image_url, category, created_at
		FROM products
		WHERE id = $1
	`, id).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.Category, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *PostgresStore) SearchProducts(query string) ([]models.Product, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, price, image_url, category, created_at
		FROM products
		WHERE name ILIKE $1 OR description ILIKE $1
		ORDER BY created_at DESC
	`, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanProducts(rows)
}

func (s *PostgresStore) ListByCategory(category string) ([]models.Product, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, price, image_url, category, created_at
		FROM products
		WHERE category = $1
		ORDER BY created_at DESC
	`, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanProducts(rows)
}

func scanProducts(rows *sql.Rows) ([]models.Product, error) {
	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.Category, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if products == nil {
		products = []models.Product{}
	}
	return products, rows.Err()
}
