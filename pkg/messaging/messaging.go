package messaging

import (
	"context"
)

// Message represents a message received from or sent to a queue
type Message struct {
	// ID is the unique identifier for the message
	ID string

	// Payload is the message content
	Payload []byte

	// Metadata contains additional information about the message
	Metadata map[string]string

	// Source identifies which queue/topic the message came from
	Source string
}

// MessageHandler represents a function that processes a message
type MessageHandler func(context.Context, *Message) error

// Consumer interface abstracts the message consuming functionality
type Consumer interface {
	// Start begins consuming messages from the queue
	Start(context.Context) error

	// Stop gracefully stops consuming messages
	Stop() error

	// Subscribe registers a handler for processing messages
	Subscribe(handler MessageHandler) error

	// Name returns the name of the consumer implementation
	Name() string
}

// Producer interface abstracts the message producing functionality
type Producer interface {
	// Publish sends a message to the queue
	Publish(ctx context.Context, topic string, message []byte, metadata map[string]string) error

	// Close gracefully closes the producer
	Close() error

	// Name returns the name of the producer implementation
	Name() string
}

// ConsumerFactory creates a new consumer
type ConsumerFactory func(config map[string]interface{}) (Consumer, error)

// ProducerFactory creates a new producer
type ProducerFactory func(config map[string]interface{}) (Producer, error)

// Registry keeps track of available messaging implementations
type Registry struct {
	consumerFactories map[string]ConsumerFactory
	producerFactories map[string]ProducerFactory
}

// NewRegistry creates a new messaging registry
func NewRegistry() *Registry {
	return &Registry{
		consumerFactories: make(map[string]ConsumerFactory),
		producerFactories: make(map[string]ProducerFactory),
	}
}

// RegisterConsumerFactory registers a consumer factory
func (r *Registry) RegisterConsumerFactory(name string, factory ConsumerFactory) {
	r.consumerFactories[name] = factory
}

// RegisterProducerFactory registers a producer factory
func (r *Registry) RegisterProducerFactory(name string, factory ProducerFactory) {
	r.producerFactories[name] = factory
}

// CreateConsumer creates a consumer with the specified implementation name
func (r *Registry) CreateConsumer(name string, config map[string]interface{}) (Consumer, error) {
	factory, ok := r.consumerFactories[name]
	if !ok {
		return nil, ErrConsumerNotFound
	}
	return factory(config)
}

// CreateProducer creates a producer with the specified implementation name
func (r *Registry) CreateProducer(name string, config map[string]interface{}) (Producer, error) {
	factory, ok := r.producerFactories[name]
	if !ok {
		return nil, ErrProducerNotFound
	}
	return factory(config)
}
