package port

import (
	"context"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/session/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// SessionRepository defines the persistence contract for sessions.
type SessionRepository interface {
	Save(ctx context.Context, session entity.Session) error
	Load(ctx context.Context, id identity.SessionID) (entity.Session, error)
}
