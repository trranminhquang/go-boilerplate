package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/trranminhquang/go-boilerplate/pkg/messaging"
)

// Job represents a task to be executed by a worker
type Job interface {
	Execute() error
	ID() string
}

// Pool represents a worker pool that manages concurrent execution of jobs
type Pool struct {
	jobQueue    chan Job
	workerCount int
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	logger      *logrus.Logger
}

// NewPool creates a new worker pool with the specified number of workers
func NewPool(workerCount int, queueSize int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())

	return &Pool{
		jobQueue:    make(chan Job, queueSize),
		workerCount: workerCount,
		ctx:         ctx,
		cancel:      cancel,
		logger:      logrus.StandardLogger(),
	}
}

// Start initializes and starts the worker pool
func (p *Pool) Start() {
	p.logger.Infof("Starting worker pool with %d workers", p.workerCount)

	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		workerID := i

		go func() {
			defer p.wg.Done()
			p.worker(workerID)
		}()
	}
}

// worker is the main worker loop that processes jobs
func (p *Pool) worker(id int) {
	p.logger.Infof("Worker %d started", id)

	for {
		select {
		case <-p.ctx.Done():
			p.logger.Infof("Worker %d shutting down", id)
			return
		case job, ok := <-p.jobQueue:
			if !ok {
				p.logger.Infof("Worker %d shutting down, job queue closed", id)
				return
			}

			p.logger.Infof("Worker %d processing job %s", id, job.ID())
			err := job.Execute()
			if err != nil {
				p.logger.Errorf("Worker %d failed to process job %s: %v", id, job.ID(), err)
			} else {
				p.logger.Infof("Worker %d completed job %s successfully", id, job.ID())
			}
		}
	}
}

// Submit adds a job to the queue
func (p *Pool) Submit(job Job) {
	select {
	case p.jobQueue <- job:
		p.logger.Infof("Job %s submitted to queue", job.ID())
	case <-p.ctx.Done():
		p.logger.Warnf("Could not submit job %s: worker pool is shutting down", job.ID())
	}
}

// Stop gracefully shuts down the worker pool
func (p *Pool) Stop() {
	p.logger.Info("Stopping worker pool")
	p.cancel()
	close(p.jobQueue)
	p.wg.Wait()
	p.logger.Info("Worker pool stopped")
}

// MessageJob is a wrapper that converts a Message to a Job
type MessageJob struct {
	message  *messaging.Message
	handler  messaging.MessageHandler
	id       string
	received time.Time
}

// NewMessageJob creates a new job from a message
func NewMessageJob(message *messaging.Message, handler messaging.MessageHandler) *MessageJob {
	return &MessageJob{
		message:  message,
		handler:  handler,
		id:       uuid.New().String(),
		received: time.Now(),
	}
}

// Execute processes the message
func (j *MessageJob) Execute() error {
	return j.handler(context.Background(), j.message)
}

// ID returns the job's unique identifier
func (j *MessageJob) ID() string {
	return j.id
}

// QueueWorker integrates the worker pool with message queues
type QueueWorker struct {
	pool     *Pool
	consumer messaging.Consumer
	logger   *logrus.Logger
	handlers messaging.HandlerRegistry
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

// WorkerOptions contains options for creating a queue worker
type WorkerOptions struct {
	ConsumerType   string
	ConsumerConfig map[string]interface{}
	WorkerCount    int
	QueueSize      int
}

// NewQueueWorker creates a new queue worker
func NewQueueWorker(options WorkerOptions, registry *messaging.Registry) (*QueueWorker, error) {
	// Create consumer
	consumer, err := registry.CreateConsumer(options.ConsumerType, options.ConsumerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	// Create worker pool
	pool := NewPool(options.WorkerCount, options.QueueSize)

	// Create context
	ctx, cancel := context.WithCancel(context.Background())

	return &QueueWorker{
		pool:     pool,
		consumer: consumer,
		logger:   logrus.StandardLogger(),
		handlers: make(messaging.HandlerRegistry),
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// RegisterHandler registers a handler for messages of the specified type
func (w *QueueWorker) RegisterHandler(msgType messaging.MessageType, handler messaging.MessageHandler) {
	w.handlers[msgType] = handler
}

// handleMessage is the default message handler
func (w *QueueWorker) handleMessage(ctx context.Context, msg *messaging.Message) error {
	w.logger.WithFields(logrus.Fields{
		"message_id": msg.ID,
		"source":     msg.Source,
	}).Info("Processing message")

	// Try to parse the message payload
	var data map[string]interface{}
	if err := json.Unmarshal(msg.Payload, &data); err != nil {
		w.logger.WithError(err).Warn("Failed to parse message payload")
	} else {
		// Try to determine message type
		var typeStr string
		if eventType, ok := data["event_type"].(string); ok {
			typeStr = eventType
		} else if msgType, ok := data["type"].(string); ok {
			typeStr = msgType
		}

		// Find a handler for the message type
		if typeStr != "" {
			msgType := messaging.MessageType(typeStr)
			if handler, ok := w.handlers[msgType]; ok {
				return handler(ctx, msg)
			}
		}
	}

	// Use default processing
	w.logger.WithFields(logrus.Fields{
		"message_id": msg.ID,
		"payload":    string(msg.Payload),
	}).Info("Processed message with default handler")

	return nil
}

// Start starts the queue worker
func (w *QueueWorker) Start() error {
	// Start the worker pool
	w.pool.Start()

	// Subscribe to consumer
	err := w.consumer.Subscribe(func(ctx context.Context, msg *messaging.Message) error {
		// Create a job from the message
		job := NewMessageJob(msg, w.handleMessage)

		// Submit the job to the worker pool
		w.pool.Submit(job)

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to consumer: %w", err)
	}

	// Start the consumer
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		err := w.consumer.Start(w.ctx)
		if err != nil {
			w.logger.WithError(err).Error("Consumer failed to start")
		}
	}()

	return nil
}

// Stop stops the queue worker
func (w *QueueWorker) Stop() error {
	w.logger.Info("Stopping queue worker")
	w.cancel()

	// Stop the consumer
	err := w.consumer.Stop()
	if err != nil {
		w.logger.WithError(err).Error("Failed to stop consumer")
	}

	// Stop the worker pool
	w.pool.Stop()

	// Wait for everything to finish
	w.wg.Wait()

	w.logger.Info("Queue worker stopped")
	return nil
}
