package kafka

import (
	"encoding/json"
	"strings"

	"github.com/IBM/sarama"
	"github.com/metalbear-co/metalmart/services/order/models"
)

type Producer struct {
	producer sarama.SyncProducer
}

func NewProducer(brokers string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	brokerList := strings.Split(brokers, ",")
	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		return nil, err
	}

	return &Producer{producer: producer}, nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}

func (p *Producer) PublishOrderCreated(event models.OrderCreatedEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: "order.created",
		Key:   sarama.StringEncoder(event.OrderID),
		Value: sarama.ByteEncoder(data),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("customer_email"),
				Value: []byte(event.CustomerEmail),
			},
		},
	}

	_, _, err = p.producer.SendMessage(msg)
	return err
}
