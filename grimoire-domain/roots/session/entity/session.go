package entity

import (
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// sessionCore holds the minimum data needed to enforce session invariants.
type sessionCore struct {
	id         identity.SessionID
	campaignID identity.CampaignID
	notes      string
}

func (c *sessionCore) SessionID() identity.SessionID   { return c.id }
func (c *sessionCore) CampaignID() identity.CampaignID { return c.campaignID }

// --- state structs ---

type newSession struct{ sessionCore }
type inProgressSession struct{ sessionCore }
type completedSession struct{ sessionCore }
type summarizedSession struct{ sessionCore }
type idleSession struct{ sessionCore }

// isSession seals the Session interface to this package.
func (*newSession) isSession()        {}
func (*inProgressSession) isSession() {}
func (*completedSession) isSession()  {}
func (*summarizedSession) isSession() {}
func (*idleSession) isSession()       {}

// Snapshot returns a reconstitutable snapshot for each state.
func (s *newSession) Snapshot() SessionSnapshot {
	return SessionSnapshot{ID: s.id, CampaignID: s.campaignID, State: "New", Notes: s.notes}
}

func (s *inProgressSession) Snapshot() SessionSnapshot {
	return SessionSnapshot{ID: s.id, CampaignID: s.campaignID, State: "InProgress", Notes: s.notes}
}

func (s *completedSession) Snapshot() SessionSnapshot {
	return SessionSnapshot{ID: s.id, CampaignID: s.campaignID, State: "Completed", Notes: s.notes}
}

func (s *summarizedSession) Snapshot() SessionSnapshot {
	return SessionSnapshot{ID: s.id, CampaignID: s.campaignID, State: "Summarized", Notes: s.notes}
}

func (s *idleSession) Snapshot() SessionSnapshot {
	return SessionSnapshot{ID: s.id, CampaignID: s.campaignID, State: "Idle", Notes: s.notes}
}

// CreateSession constructs a new session in the New state.
func CreateSession(id identity.SessionID, campaignID identity.CampaignID) (NewSession, []event.Event, error) {
	if id.IsZero() {
		return nil, nil, ErrSessionIDRequired
	}
	if campaignID.IsZero() {
		return nil, nil, ErrCampaignIDRequired
	}
	evt := event.EntityCreated{
		EntityID:   id.String(),
		EntityType: "session",
		Name:       "",
		Source:     event.SourceGrimoire,
	}
	return &newSession{sessionCore{id: id, campaignID: campaignID}}, []event.Event{evt}, nil
}
