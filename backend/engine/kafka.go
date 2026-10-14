package engine

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type CheckMessage struct {
	MonitorID      string    `json:"monitor_id"`
	URL            string    `json:"url"`
	ExpectedStatus int       `json:"expected_status"`
	ScheduledFor   time.Time `json:"scheduled_for"`
}

// EnsureTopic creates the Kafka topic if it doesn't already exist
func EnsureTopic(brokers []string, topic string, numPartitions int, replicationFactor int) error {
	if len(brokers) == 0 {
		return fmt.Errorf("no kafka brokers provided")
	}

	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer controllerConn.Close()

	topicConfig := kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
		ConfigEntries: []kafka.ConfigEntry{
			{
				ConfigName:  "retention.ms",
				ConfigValue: "3600000", // 1 hour retention
			},
		},
	}

	err = controllerConn.CreateTopics(topicConfig)
	if err != nil {
		log.Printf("Topic %s may already exist: %v", topic, err)
	} else {
		log.Printf("Created topic %s with %d partitions and 1h retention", topic, numPartitions)
	}

	return nil
}

// NewKafkaWriter creates a thread-safe Kafka writer/producer
func NewKafkaWriter(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.Hash{},
		WriteTimeout:           10 * time.Second,
		RequiredAcks:           kafka.RequireOne,
		AllowAutoTopicCreation: true,
	}
}
