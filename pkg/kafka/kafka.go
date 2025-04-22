package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/trranminhquang/go-boilerplate/pkg/messaging"
)

// Consumer implements the messaging.Consumer interface for Kafka
type Consumer struct {
	topics        []string
	brokers       []string
	groupID       string
	handler       messaging.MessageHandler
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	logger        *logrus.Logger
	consumerReady chan bool
}

// ConsumerConfig defines configuration for Kafka consumer
type ConsumerConfig struct {
	Brokers       []string
	Topics        []string
	GroupID       string
	InitialOffset string // "earliest" or "latest"
}

// NewConsumer creates a new Kafka consumer
// Note: In a real implementation, you would use an actual Kafka client library
// like github.com/Shopify/sarama or github.com/confluentinc/confluent-kafka-go
func NewConsumer(config map[string]interface{}) (messaging.Consumer, error) {
	// Parse configuration
	cfg := ConsumerConfig{}

	// Extract brokers
	if brokers, ok := config["brokers"].([]string); ok {
		cfg.Brokers = brokers
	} else if brokersStr, ok := config["brokers"].(string); ok {
		cfg.Brokers = strings.Split(brokersStr, ",")
	} else {
		return nil, fmt.Errorf("%w: brokers must be a string or []string", messaging.ErrInvalidConfig)
	}

	// Extract topics
	if topics, ok := config["topics"].([]string); ok {
		cfg.Topics = topics
	} else if topicsStr, ok := config["topics"].(string); ok {
		cfg.Topics = strings.Split(topicsStr, ",")
	} else {
		return nil, fmt.Errorf("%w: topics must be a string or []string", messaging.ErrInvalidConfig)
	}

	// Extract group ID
	if groupID, ok := config["group_id"].(string); ok {
		cfg.GroupID = groupID
	} else {
		return nil, fmt.Errorf("%w: group_id must be a string", messaging.ErrInvalidConfig)
	}

	// Create context
	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		topics:        cfg.Topics,
		brokers:       cfg.Brokers,
		groupID:       cfg.GroupID,
		ctx:           ctx,
		cancel:        cancel,
		logger:        logrus.StandardLogger(),
		consumerReady: make(chan bool),
	}, nil
}

// Start begins consuming messages from Kafka
func (c *Consumer) Start(ctx context.Context) error {
	if c.handler == nil {
		return messaging.ErrHandlerNotSet
	}

	c.logger.WithFields(logrus.Fields{
		"brokers": c.brokers,
		"topics":  c.topics,
		"groupID": c.groupID,
	}).Info("Starting Kafka consumer")

	// In a real implementation, you would configure a Kafka client here

	// Start a goroutine to handle message consumption
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		// Signal that the consumer is ready
		close(c.consumerReady)

		c.logger.Info("Kafka consumer is ready")

		// Simulated message processing loop
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				c.logger.Info("Context canceled, stopping Kafka consumer")
				return
			case <-ticker.C:
				// Simulate receiving a message
				msg := &messaging.Message{
					ID:      fmt.Sprintf("msg-%d", time.Now().Unix()),
					Payload: []byte(`{"event_type": "test_event", "data": {"key": "value"}}`),
					Source:  fmt.Sprintf("%s-%d", c.topics[0], 0),
					Metadata: map[string]string{
						"topic":     c.topics[0],
						"partition": "0",
						"offset":    fmt.Sprintf("%d", time.Now().Unix()),
					},
				}

				// Process the message
				err := c.handler(ctx, msg)
				if err != nil {
					c.logger.WithError(err).WithFields(logrus.Fields{
						"messageID": msg.ID,
						"source":    msg.Source,
					}).Error("Failed to process message")
				} else {
					c.logger.WithFields(logrus.Fields{
						"messageID": msg.ID,
						"source":    msg.Source,
					}).Info("Successfully processed message")
				}
			}
		}
	}()

	// Wait for the consumer to be ready
	<-c.consumerReady
	return nil
}

// Stop gracefully stops consuming messages
func (c *Consumer) Stop() error {
	c.logger.Info("Stopping Kafka consumer")
	c.cancel()
	c.wg.Wait()
	c.logger.Info("Kafka consumer stopped")
	return nil
}

// Subscribe registers a handler for processing messages
func (c *Consumer) Subscribe(handler messaging.MessageHandler) error {
	c.handler = handler
	return nil
}

// Name returns the name of the consumer implementation
func (c *Consumer) Name() string {
	return "kafka"
}

// Producer implements the messaging.Producer interface for Kafka
type Producer struct {
	brokers []string
	logger  *logrus.Logger
	mu      sync.Mutex
}

// ProducerConfig defines configuration for Kafka producer
type ProducerConfig struct {
	Brokers []string
}

// NewProducer creates a new Kafka producer
// Note: In a real implementation, you would use an actual Kafka client library
func NewProducer(config map[string]interface{}) (messaging.Producer, error) {
	// Parse configuration
	cfg := ProducerConfig{}

	// Extract brokers
	if brokers, ok := config["brokers"].([]string); ok {
		cfg.Brokers = brokers
	} else if brokersStr, ok := config["brokers"].(string); ok {
		cfg.Brokers = strings.Split(brokersStr, ",")
	} else {
		return nil, fmt.Errorf("%w: brokers must be a string or []string", messaging.ErrInvalidConfig)
	}

	return &Producer{
		brokers: cfg.Brokers,
		logger:  logrus.StandardLogger(),
	}, nil
}

// Publish sends a message to Kafka
func (p *Producer) Publish(ctx context.Context, topic string, message []byte, metadata map[string]string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.logger.WithFields(logrus.Fields{
		"topic":     topic,
		"messageID": metadata["messageID"],
		"brokers":   p.brokers,
	}).Info("Publishing message to Kafka")

	// In a real implementation, you would send the message to Kafka here

	// Log the message for demonstration purposes
	var payload interface{}
	if err := json.Unmarshal(message, &payload); err != nil {
		p.logger.WithError(err).Warn("Failed to parse message payload")
		payload = string(message)
	}

	p.logger.WithFields(logrus.Fields{
		"topic":   topic,
		"payload": payload,
	}).Debug("Message published")

	return nil
}

// Close gracefully closes the producer
func (p *Producer) Close() error {
	p.logger.Info("Closing Kafka producer")
	// In a real implementation, you would close the Kafka producer here
	return nil
}

// Name returns the name of the producer implementation
func (p *Producer) Name() string {
	return "kafka"
}

// Register registers Kafka implementations with the registry
func Register(registry *messaging.Registry) {
	registry.RegisterConsumerFactory("kafka", NewConsumer)
	registry.RegisterProducerFactory("kafka", NewProducer)
}
