package kafka

import (
	"company-service/models"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"sync"
)

type Producer interface {
	ProduceEvent(event *EventMessage) error
	Close()
}

// Producer wraps the Kafka producer
type KafkaProducer struct {
	producer *kafka.Producer
	topic    string
	wg       sync.WaitGroup
}
type EventMessage struct {
	EventType string          `json:"event_type"`
	Timestamp string          `json:"timestamp"`
	Company   *models.Company `json:"company"`
}

// InitializeKafkaProducer initializes a Kafka producer
func NewKafkaProducer(kafkaURL string, topic string) (*KafkaProducer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaURL,
	})
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{producer: p, topic: topic}, nil
}

func (p *KafkaProducer) ProduceEvent(event *EventMessage) error {
	deliveryChan := make(chan kafka.Event)
	// Serialize the event to JSON
	eventBytes, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return err
	}

	// Produce the event
	err = p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
		Value:          eventBytes,
	}, deliveryChan)

	if err != nil {
		return err
	}

	p.wg.Add(1)
	// Handle the delivery report asynchronously
	go func() {
		defer func() {
			close(deliveryChan)
			p.wg.Done()
		}()
		ev := <-deliveryChan
		switch e := ev.(type) {
		case *kafka.Message:
			if e.TopicPartition.Error != nil {
				log.Printf("Error producing message: %v", e.TopicPartition.Error)
			} else {
				log.Printf("Produced event: %s", e.Value)
			}
		}
	}()

	return nil
}

// Close the producer when done
func (p *KafkaProducer) Close() {
	p.wg.Wait()
	p.producer.Close()
}
