package event_test

import (
	"testing"
	"time"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

func TestEntityCreated_ImplementsEvent(t *testing.T) {
	var e event.Event = event.EntityCreated{
		EntityID:   "id-1",
		EntityType: "game",
		Name:       "My Game",
		Source:     event.SourceGrimoire,
	}
	if e.EventType() != event.TypeEntityCreated {
		t.Errorf("expected %s, got %s", event.TypeEntityCreated, e.EventType())
	}
}

func TestEntityUpdated_ImplementsEvent(t *testing.T) {
	var e event.Event = event.EntityUpdated{
		EntityID: "id-1",
		Field:    "name",
		OldValue: "old",
		NewValue: "new",
		Source:   event.SourceObsidian,
	}
	if e.EventType() != event.TypeEntityUpdated {
		t.Errorf("expected %s, got %s", event.TypeEntityUpdated, e.EventType())
	}
}

func TestEntityLinked_ImplementsEvent(t *testing.T) {
	var e event.Event = event.EntityLinked{
		EntityAID:    "id-a",
		EntityBID:    "id-b",
		Relationship: "campaign",
		Source:       event.SourceFoundry,
	}
	if e.EventType() != event.TypeEntityLinked {
		t.Errorf("expected %s, got %s", event.TypeEntityLinked, e.EventType())
	}
}

func TestEntityRevealed_ImplementsEvent(t *testing.T) {
	var e event.Event = event.EntityRevealed{
		EntityID:   "id-1",
		RevealedTo: []string{"player-1", "player-2"},
		SessionID:  "session-1",
		Source:     event.SourceGrimoire,
	}
	if e.EventType() != event.TypeEntityRevealed {
		t.Errorf("expected %s, got %s", event.TypeEntityRevealed, e.EventType())
	}
}

func TestSessionStarted_ImplementsEvent(t *testing.T) {
	var e event.Event = event.SessionStarted{
		SessionID:  "session-1",
		CampaignID: "campaign-1",
		Date:       time.Now(),
	}
	if e.EventType() != event.TypeSessionStarted {
		t.Errorf("expected %s, got %s", event.TypeSessionStarted, e.EventType())
	}
}

func TestSessionEnded_ImplementsEvent(t *testing.T) {
	var e event.Event = event.SessionEnded{
		SessionID: "session-1",
		Notes:     "Great session",
	}
	if e.EventType() != event.TypeSessionEnded {
		t.Errorf("expected %s, got %s", event.TypeSessionEnded, e.EventType())
	}
}

func TestEventEnvelope_CarriesTypedPayload(t *testing.T) {
	eventID, _ := identity.ParseEventID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	gameID := identity.NewGameID()
	campID := identity.NewCampaignID()

	env := event.EventEnvelope{
		ID:             eventID,
		Type:           event.TypeEntityCreated,
		AggregateID:    gameID.GrimoireID,
		AggregateType:  event.AggregateGame,
		SequenceNumber: 1,
		CampaignID:     campID,
		Source:         event.SourceGrimoire,
		ActorID:        "gm-1",
		OccurredAt:     time.Now(),
		Payload: event.EntityCreated{
			EntityID:   gameID.String(),
			EntityType: "game",
			Name:       "My Game",
			Source:     event.SourceGrimoire,
		},
	}

	if env.Payload.EventType() != event.TypeEntityCreated {
		t.Errorf("expected payload type %s, got %s", event.TypeEntityCreated, env.Payload.EventType())
	}
	if env.AggregateType != event.AggregateGame {
		t.Errorf("expected aggregate type 'game', got %q", env.AggregateType)
	}
	if env.SequenceNumber != 1 {
		t.Errorf("expected sequence number 1, got %d", env.SequenceNumber)
	}
	if env.ID.String() != "01ARZ3NDEKTSV4RRFFQ69G5FAV" {
		t.Errorf("expected event ID '01ARZ3NDEKTSV4RRFFQ69G5FAV', got %q", env.ID.String())
	}
}

func TestEventEnvelope_ZeroSessionID_WhenNotApplicable(t *testing.T) {
	env := event.EventEnvelope{
		Type:          event.TypeEntityCreated,
		AggregateType: event.AggregateGame,
		Payload: event.EntityCreated{
			EntityID:   "id-1",
			EntityType: "game",
			Name:       "My Game",
			Source:     event.SourceGrimoire,
		},
	}

	if !env.SessionID.IsZero() {
		t.Error("expected zero SessionID when not applicable")
	}
	if !env.CampaignID.IsZero() {
		t.Error("expected zero CampaignID when not applicable")
	}
}
