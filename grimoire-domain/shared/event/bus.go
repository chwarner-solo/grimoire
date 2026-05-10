package event

import "context"

// EventHandler processes a dispatched event envelope.
type EventHandler interface {
	Handle(ctx context.Context, envelope EventEnvelope) error
}

// EventBus dispatches event envelopes to registered handlers.
type EventBus interface {
	// Dispatch sends an event envelope to all registered handlers for its type.
	Dispatch(ctx context.Context, envelope EventEnvelope) error

	// Subscribe registers a handler for a specific event type.
	Subscribe(eventType EventType, handler EventHandler) error
}
