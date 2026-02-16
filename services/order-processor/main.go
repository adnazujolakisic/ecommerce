package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

type OrderCreatedEvent struct {
	OrderID       string    `json:"order_id"`
	OrderNumber   string    `json:"order_number"`
	CustomerEmail string    `json:"customer_email"`
	TotalAmount   float64   `json:"total_amount"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type OrderProcessor struct {
	orderURL   string
	httpClient *http.Client
}

func main() {
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	orderURL := os.Getenv("ORDER_SERVICE_URL")
	if orderURL == "" {
		orderURL = "http://localhost:8084"
	}

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "order.created"
	}

	kafkaGroupID := os.Getenv("KAFKA_GROUP_ID")
	if kafkaGroupID == "" {
		kafkaGroupID = "order-processor"
	}

	processor := &OrderProcessor{
		orderURL:   orderURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	brokerList := strings.Split(kafkaBrokers, ",")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumerGroup, err := sarama.NewConsumerGroup(brokerList, kafkaGroupID, config)
	if err != nil {
		log.Fatalf("Failed to create consumer group: %v", err)
	}
	defer consumerGroup.Close()

	handler := &ConsumerHandler{processor: processor}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := consumerGroup.Consume(ctx, []string{kafkaTopic}, handler); err != nil {
					log.Printf("Error consuming: %v", err)
					time.Sleep(5 * time.Second)
				}
			}
		}
	}()

	log.Println("Order processor started, waiting for messages...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	log.Println("Shutting down order processor...")
	cancel()
}

type ConsumerHandler struct {
	processor *OrderProcessor
}

func (h *ConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *ConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		log.Printf("Received message: topic=%s partition=%d offset=%d", msg.Topic, msg.Partition, msg.Offset)

		var event OrderCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			session.MarkMessage(msg, "")
			continue
		}

		if err := h.processor.ProcessOrder(event, msg.Topic); err != nil {
			log.Printf("Failed to process order %s: %v", event.OrderID, err)
		}

		session.MarkMessage(msg, "")
	}
	return nil
}

func (p *OrderProcessor) ProcessOrder(event OrderCreatedEvent, kafkaTopic string) error {
	log.Printf("Processing order %s (order number: %s) [Kafka topic: %s]", event.OrderID, event.OrderNumber, kafkaTopic)

	// Demo mode: faster processing for load testing
	// Set DEMO_MODE=true to reduce delays
	demoMode := os.Getenv("DEMO_MODE") == "true"
	
	var steps []struct {
		status string
		delay  time.Duration
	}
	
	if demoMode {
		// Fast processing for demo: 200ms per step
		steps = []struct {
			status string
			delay  time.Duration
		}{
			{"processing", 200 * time.Millisecond},
			{"confirmed", 200 * time.Millisecond},
			{"shipped", 200 * time.Millisecond},
		}
	} else {
		// Normal processing
		steps = []struct {
			status string
			delay  time.Duration
		}{
			{"processing", 2 * time.Second},
			{"confirmed", 3 * time.Second},
			{"shipped", 5 * time.Second},
		}
	}

	for _, step := range steps {
		time.Sleep(step.delay)

		if err := p.updateOrderStatus(event.OrderID, step.status, kafkaTopic); err != nil {
			log.Printf("Failed to update order %s to status %s: %v", event.OrderID, step.status, err)
			return err
		}

		log.Printf("Order %s status updated to: %s", event.OrderNumber, step.status)
	}

	log.Printf("Order %s processing complete", event.OrderNumber)
	return nil
}

func (p *OrderProcessor) updateOrderStatus(orderID, status, kafkaTopic string) error {
	body, _ := json.Marshal(map[string]string{"status": status})

	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/api/orders/%s/status", p.orderURL, orderID),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	// Signal to Order Service that this update came from Kafka (mirrord queue splitting)
	req.Header.Set("X-Processor-Source", "mirrord-kafka")
	req.Header.Set("X-Kafka-Topic", kafkaTopic)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
