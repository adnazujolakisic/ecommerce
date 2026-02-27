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
		receivedAt := time.Now().Format("15:04:05.000")
		log.Printf("[%s] Received message (status still pending): topic=%s partition=%d offset=%d",
			receivedAt, msg.Topic, msg.Partition, msg.Offset)

		var event OrderCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			session.MarkMessage(msg, "")
			continue
		}

		if b, _ := json.MarshalIndent(event, "", "  "); len(b) > 0 {
			log.Printf("Received from Kafka: %s", string(b))
		}

		log.Printf("Processing order %s (customer=%s) - will update status pending→processing→confirmed→shipped",
			event.OrderNumber, event.CustomerEmail)
		if err := h.processor.ProcessOrder(event, msg.Topic); err != nil {
			log.Printf("Failed to process order %s: %v", event.OrderID, err)
		}

		session.MarkMessage(msg, "")
	}
	return nil
}

func (p *OrderProcessor) ProcessOrder(event OrderCreatedEvent, kafkaTopic string) error {
	log.Printf("Processing order %s (order number: %s) [Kafka topic: %s]", event.OrderID, event.OrderNumber, kafkaTopic)

	steps := []struct {
		status string
		delay  time.Duration
	}{
		{"processing", 2 * time.Second},
		{"confirmed", 2 * time.Second},
		{"shipped", 2 * time.Second},
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
	// Only signal mirrord when consuming from mirrord's temp topic (queue splitting)
	if strings.Contains(kafkaTopic, "mirrord-tmp") {
		req.Header.Set("X-Processor-Source", "mirrord-kafka")
		req.Header.Set("X-Kafka-Topic", kafkaTopic)
	}

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
