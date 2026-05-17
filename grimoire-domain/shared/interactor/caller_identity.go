package interactor

import "context"

// CallerIdentityPort verifies a caller token and returns the caller's ID.
// The interactor layer trusts the returned ID completely — authentication
// is performed by the adapter before the result reaches any interactor.
type CallerIdentityPort interface {
	Identify(ctx context.Context, token string) (string, error)
}