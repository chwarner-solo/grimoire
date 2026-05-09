package event

import "time"

// Event is the sealed interface for all canonical event payloads.
// The unexported isEvent() marker prevents external implementations.
type Event interface {
	isEvent()
	EventType() EventType
}

// EntityCreated records the creation of a domain entity.
type EntityCreated struct {
	EntityID   string
	EntityType string
	Name       string
	Source     Source
}

func (EntityCreated) isEvent()              {}
func (EntityCreated) EventType() EventType  { return TypeEntityCreated }

// EntityUpdated records a field-level change to an existing entity.
type EntityUpdated struct {
	EntityID string
	Field    string
	OldValue string
	NewValue string
	Source   Source
}

func (EntityUpdated) isEvent()              {}
func (EntityUpdated) EventType() EventType  { return TypeEntityUpdated }

// EntityLinked records a relationship between two entities.
type EntityLinked struct {
	EntityAID    string
	EntityBID    string
	Relationship string
	Source       Source
}

func (EntityLinked) isEvent()              {}
func (EntityLinked) EventType() EventType  { return TypeEntityLinked }

// EntityRevealed records that an entity has been revealed to players.
type EntityRevealed struct {
	EntityID   string
	RevealedTo []string
	SessionID  string
	Source     Source
}

func (EntityRevealed) isEvent()              {}
func (EntityRevealed) EventType() EventType  { return TypeEntityRevealed }

// SessionStarted records the beginning of a play session.
type SessionStarted struct {
	SessionID  string
	CampaignID string
	Date       time.Time
}

func (SessionStarted) isEvent()              {}
func (SessionStarted) EventType() EventType  { return TypeSessionStarted }

// SessionEnded records the end of a play session.
type SessionEnded struct {
	SessionID string
	Notes     string
}

func (SessionEnded) isEvent()              {}
func (SessionEnded) EventType() EventType  { return TypeSessionEnded }
