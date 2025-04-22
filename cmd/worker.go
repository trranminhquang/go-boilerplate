package cmd

import (
	"context"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/trranminhquang/go-boilerplate/internal/worker"
	"github.com/trranminhquang/go-boilerplate/pkg/kafka"
	"github.com/trranminhquang/go-boilerplate/pkg/messaging"
)

// Worker command flags
var (
	numWorkers int
	queueSize  int
	queueType  string
	brokers    string
	topics     string
	groupID    string
)

// workerCmd represents the worker command
var workerCmd = cobra.Command{
	Use:   "worker",
	Short: "Start message queue workers",
	Long:  "Start workers to process tasks from message queues like Kafka",
	Run: func(cmd *cobra.Command, args []string) {
		startWorker(cmd.Context())
	},
}

func init() {
	defaultOpts := worker.DefaultOptions()

	workerCmd.Flags().IntVarP(&numWorkers, "workers", "w", defaultOpts.WorkerCount, "Number of concurrent workers")
	workerCmd.Flags().IntVarP(&queueSize, "queue-size", "q", defaultOpts.QueueSize, "Maximum size of the job queue")
	workerCmd.Flags().StringVarP(&queueType, "queue-type", "t", defaultOpts.QueueType, "Type of message queue to use (kafka)")
	workerCmd.Flags().StringVarP(&brokers, "brokers", "b", defaultOpts.QueueConfig["brokers"].(string), "Comma-separated list of message queue brokers")
	workerCmd.Flags().StringVarP(&topics, "topics", "", defaultOpts.QueueConfig["topics"].(string), "Comma-separated list of topics to consume")
	workerCmd.Flags().StringVarP(&groupID, "group-id", "g", defaultOpts.QueueConfig["group_id"].(string), "Consumer group ID")
}

// startWorker initializes and runs the worker process
func startWorker(ctx context.Context) {
	logrus.Info("Starting worker with concurrency: ", numWorkers)

	// Create messaging registry
	registry := messaging.NewRegistry()

	// Register Kafka implementation
	kafka.Register(registry)

	// Create queue configuration based on command line flags
	queueConfig := map[string]interface{}{
		"brokers":  strings.Split(brokers, ","),
		"topics":   strings.Split(topics, ","),
		"group_id": groupID,
	}

	// Create queue worker options
	options := worker.WorkerOptions{
		ConsumerType:   queueType,
		ConsumerConfig: queueConfig,
		WorkerCount:    numWorkers,
		QueueSize:      queueSize,
	}

	// Create queue worker
	queueWorker, err := worker.NewQueueWorker(options, registry)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create queue worker")
	}

	// Register message handlers
	registerMessageHandlers(queueWorker)

	// Start the worker
	if err := queueWorker.Start(); err != nil {
		logrus.WithError(err).Fatal("Failed to start queue worker")
	}

	logrus.Info("Worker started and consuming messages")

	// Wait for context cancellation (CTRL+C or shutdown signal)
	<-ctx.Done()
	logrus.Info("Shutting down worker...")

	// Stop the worker
	if err := queueWorker.Stop(); err != nil {
		logrus.WithError(err).Error("Error stopping worker")
	}

	logrus.Info("Worker has been shut down")
	os.Exit(0)
}

// registerMessageHandlers registers handlers for different message types
func registerMessageHandlers(queueWorker *worker.QueueWorker) {
	// User-related message handlers
	queueWorker.RegisterHandler(messaging.UserCreated, handleUserCreated)

	// Order-related message handlers
	queueWorker.RegisterHandler(messaging.OrderPlaced, handleOrderPlaced)

	// Payment-related message handlers
	queueWorker.RegisterHandler(messaging.PaymentReceived, handlePaymentReceived)

	// Notification-related message handlers
	queueWorker.RegisterHandler(messaging.NotificationSent, handleNotificationSent)
}

// Message handler functions
func handleUserCreated(ctx context.Context, msg *messaging.Message) error {
	logrus.WithField("messageID", msg.ID).Info("Handling user created message")
	// Process user creation message
	return nil
}

func handleOrderPlaced(ctx context.Context, msg *messaging.Message) error {
	logrus.WithField("messageID", msg.ID).Info("Handling order placed message")
	// Process order placement message
	return nil
}

func handlePaymentReceived(ctx context.Context, msg *messaging.Message) error {
	logrus.WithField("messageID", msg.ID).Info("Handling payment received message")
	// Process payment message
	return nil
}

func handleNotificationSent(ctx context.Context, msg *messaging.Message) error {
	logrus.WithField("messageID", msg.ID).Info("Handling notification sent message")
	// Process notification message
	return nil
}
