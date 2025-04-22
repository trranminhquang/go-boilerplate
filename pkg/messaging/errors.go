package messaging

import "errors"

// Error definitions for messaging
var (
	ErrConsumerNotFound = errors.New("consumer implementation not found")
	ErrProducerNotFound = errors.New("producer implementation not found")
	ErrHandlerNotSet    = errors.New("message handler not set")
	ErrInvalidConfig    = errors.New("invalid configuration")
)
