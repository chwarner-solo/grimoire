package entity

import (
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// Session is the sealed base interface for all session states.
type Session interface {
	isSession()
	SessionID() identity.SessionID
	CampaignID() identity.CampaignID
	Snapshot() SessionSnapshot
	Handle(evt event.Event) (Session, error)
}

// NewSession represents a session that has been created but not yet started.
type NewSession interface {
	Session
	Start() (InProgressSession, error)
}

// InProgressSession represents a session currently being played.
type InProgressSession interface {
	Session
	End() (CompletedSession, error)
}

// CompletedSession represents a session that has ended but not yet summarized.
type CompletedSession interface {
	Session
	Summarize(notes string) (SummarizedSession, error)
}

// SummarizedSession represents a session whose recap has been written.
type SummarizedSession interface {
	Session
	Close() (IdleSession, error)
}

// IdleSession is the terminal state for a session.
type IdleSession interface {
	Session
}
