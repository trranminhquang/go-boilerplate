package messaging

// MessageType represents the type of a message
type MessageType string

// String returns the string representation of the message type
func (mt MessageType) String() string {
	return string(mt)
}

// Message type constants
const (
	// User related message types
	UserCreated MessageType = "user_created"
	UserUpdated MessageType = "user_updated"
	UserDeleted MessageType = "user_deleted"

	// Order related message types
	OrderPlaced    MessageType = "order_placed"
	OrderPaid      MessageType = "order_paid"
	OrderShipped   MessageType = "order_shipped"
	OrderDelivered MessageType = "order_delivered"
	OrderCanceled  MessageType = "order_canceled"

	// Payment related message types
	PaymentReceived MessageType = "payment_received"
	PaymentFailed   MessageType = "payment_failed"
	PaymentRefunded MessageType = "payment_refunded"

	// Notification related message types
	NotificationSent   MessageType = "notification_sent"
	NotificationFailed MessageType = "notification_failed"
)

// AllMessageTypes returns a slice of all defined message types
func AllMessageTypes() []MessageType {
	return []MessageType{
		// User related
		UserCreated,
		UserUpdated,
		UserDeleted,

		// Order related
		OrderPlaced,
		OrderPaid,
		OrderShipped,
		OrderDelivered,
		OrderCanceled,

		// Payment related
		PaymentReceived,
		PaymentFailed,
		PaymentRefunded,

		// Notification related
		NotificationSent,
		NotificationFailed,
	}
}

// HandlerRegistry maps message types to their handlers
type HandlerRegistry map[MessageType]MessageHandler
