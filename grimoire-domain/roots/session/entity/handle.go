package entity

import (
	"errors"
	"fmt"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- newSession Handle ---

func (s *newSession) Handle(evt event.Event) (Session, error) {
	switch evt.(type) {
	case event.SessionStarted:
		core := s.sessionCore
		return &inProgressSession{core}, nil
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- inProgressSession Handle ---

func (s *inProgressSession) Handle(evt event.Event) (Session, error) {
	switch evt.(type) {
	case event.SessionEnded:
		core := s.sessionCore
		return &completedSession{core}, nil
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- completedSession Handle ---

func (s *completedSession) Handle(evt event.Event) (Session, error) {
	switch e := evt.(type) {
	case event.EntityUpdated:
		if e.Field == "notes" {
			core := s.sessionCore
			core.notes = e.NewValue
			return &summarizedSession{core}, nil
		}
		return nil, ErrUnexpectedEvent
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- summarizedSession Handle ---

func (s *summarizedSession) Handle(evt event.Event) (Session, error) {
	switch e := evt.(type) {
	case event.EntityUpdated:
		if e.Field == "status" && e.NewValue == "idle" {
			core := s.sessionCore
			return &idleSession{core}, nil
		}
		return nil, ErrUnexpectedEvent
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- idleSession Handle ---

func (s *idleSession) Handle(_ event.Event) (Session, error) {
	return nil, ErrUnexpectedEvent
}

// --- ReplaySession ---

// ReplaySession rebuilds a Session aggregate from a sequence of events.
// The first event must be an EntityCreated for a session.
// The campaignID is required because it's not stored in the event payload.
func ReplaySession(campaignID identity.CampaignID, events []event.Event) (Session, error) {
	if len(events) == 0 {
		return nil, errors.New("session: no events to replay")
	}

	first, ok := events[0].(event.EntityCreated)
	if !ok {
		return nil, errors.New("session: first event must be EntityCreated")
	}

	sessID, err := identity.ParseSessionID(first.EntityID)
	if err != nil {
		return nil, fmt.Errorf("session replay: %w", err)
	}

	s, err := CreateSession(sessID, campaignID)
	if err != nil {
		return nil, fmt.Errorf("session replay: %w", err)
	}

	var current Session = s
	for _, evt := range events[1:] {
		current, err = current.Handle(evt)
		if err != nil {
			return nil, fmt.Errorf("session replay: %w", err)
		}
	}
	return current, nil
}
