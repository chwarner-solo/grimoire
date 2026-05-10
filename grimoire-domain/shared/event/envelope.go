package event

import (
	"time"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// EventEnvelope wraps an event payload with routing, sequencing, and audit metadata.
type EventEnvelope struct {
	ID             identity.EventID     // globally sortable ULID
	Type           EventType
	AggregateID    identity.GrimoireID  // which aggregate this event belongs to
	AggregateType  AggregateType        // game, campaign, session, etc.
	SequenceNumber uint64               // monotonic counter within aggregate
	CampaignID     identity.CampaignID  // routing context (zero if not applicable)
	SessionID      identity.SessionID   // routing context (zero if not applicable)
	Source         Source               // obsidian | foundry | grimoire
	ActorID        string
	OccurredAt     time.Time
	Payload        Event                // typed domain event
}
