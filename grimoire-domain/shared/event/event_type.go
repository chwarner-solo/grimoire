package event

// EventType identifies which of the six canonical event types an event is.
type EventType string

const (
	TypeEntityCreated  EventType = "EntityCreated"
	TypeEntityUpdated  EventType = "EntityUpdated"
	TypeEntityLinked   EventType = "EntityLinked"
	TypeEntityRevealed EventType = "EntityRevealed"
	TypeSessionStarted EventType = "SessionStarted"
	TypeSessionEnded   EventType = "SessionEnded"
)
