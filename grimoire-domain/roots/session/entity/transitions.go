package entity

import (
	"strings"
	"time"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
)

// Start transitions a NewSession to InProgressSession.
func (s *newSession) Start(date time.Time) (InProgressSession, []event.Event, error) {
	evt := event.SessionStarted{
		SessionID:  s.id.String(),
		CampaignID: s.campaignID.String(),
		Date:       date,
	}
	return &inProgressSession{s.sessionCore}, []event.Event{evt}, nil
}

// End transitions an InProgressSession to CompletedSession.
func (s *inProgressSession) End() (CompletedSession, []event.Event, error) {
	evt := event.SessionEnded{
		SessionID: s.id.String(),
		Notes:     "",
	}
	return &completedSession{s.sessionCore}, []event.Event{evt}, nil
}

// Summarize transitions a CompletedSession to SummarizedSession.
// Notes must be non-empty.
func (s *completedSession) Summarize(notes string, source event.Source) (SummarizedSession, []event.Event, error) {
	if strings.TrimSpace(notes) == "" {
		return nil, nil, ErrNotesRequired
	}
	core := s.sessionCore
	core.notes = notes
	evt := event.EntityUpdated{
		EntityID: s.id.String(),
		Field:    "notes",
		OldValue: "",
		NewValue: notes,
		Source:   source,
	}
	return &summarizedSession{core}, []event.Event{evt}, nil
}

// Close transitions a SummarizedSession to IdleSession (terminal).
func (s *summarizedSession) Close(source event.Source) (IdleSession, []event.Event, error) {
	evt := event.EntityUpdated{
		EntityID: s.id.String(),
		Field:    "status",
		OldValue: "summarized",
		NewValue: "idle",
		Source:   source,
	}
	return &idleSession{s.sessionCore}, []event.Event{evt}, nil
}
