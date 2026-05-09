package entity

import (
	"fmt"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// SessionSnapshot is the serialisable representation of a session's state.
type SessionSnapshot struct {
	ID         identity.SessionID
	CampaignID identity.CampaignID
	State      string
	Notes      string
}

// ReconstituteSession rebuilds a Session from a snapshot without re-validating
// creation invariants. This is used when loading from persistence.
func ReconstituteSession(snap SessionSnapshot) (Session, error) {
	core := sessionCore{
		id:         snap.ID,
		campaignID: snap.CampaignID,
		notes:      snap.Notes,
	}

	switch snap.State {
	case "New":
		return &newSession{core}, nil
	case "InProgress":
		return &inProgressSession{core}, nil
	case "Completed":
		return &completedSession{core}, nil
	case "Summarized":
		return &summarizedSession{core}, nil
	case "Idle":
		return &idleSession{core}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidSessionState, snap.State)
	}
}
