package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type OrderRequest struct {
	CustomerEmail   string      `json:"customer_email"`
	CustomerName    string      `json:"customer_name"`
	ShippingAddress Address     `json:"shipping_address"`
	Items           []OrderItem `json:"items"`
	ReservationID   string      `json:"reservation_id"`
}

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
	Country string `json:"country"`
}

type OrderItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

var products = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
var names = []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank", "Grace", "Henry"}

func main() {
	rate := 10
	duration := 60
	baseURL := "http://localhost:8084"

	if len(os.Args) > 1 {
		if r, err := strconv.Atoi(os.Args[1]); err == nil {
			rate = r
		}
	}
	if len(os.Args) > 2 {
		if d, err := strconv.Atoi(os.Args[2]); err == nil {
			duration = d
		}
	}
	if url := os.Getenv("BASE_URL"); url != "" {
		baseURL = url
	}

	fmt.Printf("🚀 MetalMart Load Generator\n")
	fmt.Printf("==========================\n")
	fmt.Printf("Rate: %d orders/second\n", rate)
	fmt.Printf("Duration: %d seconds\n", duration)
	fmt.Printf("Total orders: %d\n", rate*duration)
	fmt.Printf("Base URL: %s\n\n", baseURL)

	delay := time.Duration(1000/rate) * time.Millisecond
	startTime := time.Now()
	endTime := startTime.Add(time.Duration(duration) * time.Second)
	orderCount := 0

	client := &http.Client{Timeout: 5 * time.Second}

	fmt.Println("Starting load generation...")
	fmt.Println("Press Ctrl+C to stop early\n")

	for time.Now().Before(endTime) {
		orderCount++
		createOrder(client, baseURL, orderCount)

		if orderCount%10 == 0 {
			fmt.Printf(".")
		}

		time.Sleep(delay)
	}

	fmt.Printf("\n\n✅ Load generation complete!\n")
	fmt.Printf("Total orders created: %d\n", orderCount)
	fmt.Printf("Actual rate: %.2f orders/second\n", float64(orderCount)/float64(duration))
}

func createOrder(client *http.Client, baseURL string, orderID int) {
	productID := products[orderID%len(products)]
	name := names[orderID%len(names)]
	email := fmt.Sprintf("load-test-%d@metalbear.com", orderID)

	req := OrderRequest{
		CustomerEmail: email,
		CustomerName:  name,
		ShippingAddress: Address{
			Street:  fmt.Sprintf("%d Test St", orderID),
			City:    "Test City",
			State:   "TS",
			ZipCode: "12345",
			Country: "USA",
		},
		Items: []OrderItem{
			{
				ProductID:   productID,
				ProductName: "Test Product",
				Quantity:    (orderID % 3) + 1,
				Price:       float64((orderID%50)+10) + 0.99,
			},
		},
		ReservationID: fmt.Sprintf("res_load_%d", orderID),
	}

	body, _ := json.Marshal(req)
	resp, err := client.Post(baseURL+"/api/orders", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creating order %d: %v", orderID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Printf("Failed to create order %d: status %d", orderID, resp.StatusCode)
	}
}
