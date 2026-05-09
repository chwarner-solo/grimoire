package event

import "time"

// EventEnvelope wraps an event payload with routing and audit metadata.
type EventEnvelope struct {
	ID         string
	Type       EventType
	CampaignID string
	SessionID  string
	Source     Source
	ActorID    string
	OccurredAt time.Time
	Payload    Event
}
