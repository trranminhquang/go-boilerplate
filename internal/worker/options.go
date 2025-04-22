package worker

// Config represents configuration options for the worker pool
type Config struct {
	// WorkerCount is the number of concurrent workers to run
	WorkerCount int

	// QueueSize is the maximum number of jobs that can be queued
	QueueSize int

	// QueueType specifies the type of message queue to use (e.g., "kafka")
	QueueType string

	// QueueConfig contains configuration for the specific queue type
	QueueConfig map[string]interface{}
}

// DefaultOptions returns default worker configuration
func DefaultOptions() Config {
	return Config{
		WorkerCount: 4,
		QueueSize:   100,
		QueueType:   "kafka",
		QueueConfig: map[string]interface{}{
			"brokers":  "localhost:9092",
			"topics":   "default-topic",
			"group_id": "go-worker-group",
		},
	}
}
