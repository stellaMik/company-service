package kafka

import (
	"context"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
)

// KafkaConsumer wraps the Kafka consumer
type KafkaConsumer struct {
	consumer *kafka.Consumer
}

// NewKafkaConsumer initializes a Kafka consumer
func NewKafkaConsumer(kafkaURL, groupID, topic string) (*KafkaConsumer, error) {
	// Create a new Kafka consumer instance
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaURL,
		"group.id":          groupID,
		"auto.offset.reset": "earliest", // Start consuming from the earliest message
	})
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}

	// Subscribe to the Kafka topic
	err = c.Subscribe(topic, nil)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{consumer: c}, nil
}

// ConsumeEvents listens for events from Kafka and processes them
func (c *KafkaConsumer) ConsumeEvents() {
	for {
		// Read message with a timeout of 1000 milliseconds
		msg, err := c.consumer.ReadMessage(1000)
		if err != nil {
			if err.(kafka.Error).Code() == kafka.ErrTimedOut {
				// Timeout error is expected for polling, so we continue
				continue
			}
			log.Printf("Error consuming message: %v", err)
			continue
		}

		var event map[string]interface{}
		err = json.Unmarshal(msg.Value, &event)
		if err != nil {
			log.Printf("Error unmarshaling event: %v", err)
			continue
		}

		log.Printf("Received event: %s, Data: %v", event["event"], event["data"])
		// You can process the event here (e.g., logging, updating the database, etc.)
	}
}

// ConsumeMessage listens for messages from a specific topic and sends them to a channel
func (c *KafkaConsumer) ConsumeMessage(ctx context.Context, topic string, consumedMessages chan EventMessage) {
	for {
		select {
		case <-ctx.Done(): // Context canceled or expired
			log.Println("Context canceled. Stopping message consumption.")
			return

		default:
			msg, err := c.consumer.ReadMessage(100 * 1000) // Set timeout for consuming
			if err == nil && msg.TopicPartition.Topic != nil && *msg.TopicPartition.Topic == topic {
				var event EventMessage
				err = json.Unmarshal(msg.Value, &event)
				if err != nil {
					log.Printf("Error unmarshaling event: %v", err)
					continue
				}
				consumedMessages <- event
			}
		}
	}
}

// Close the consumer when done
func (c *KafkaConsumer) Close() error {
	err := c.consumer.Close()
	if err != nil {
		return err
	}
	return nil
}
