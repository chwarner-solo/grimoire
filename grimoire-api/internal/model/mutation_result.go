package model

// MutationResult is the uniform response type for all GM mutations.
// Callers use ID to issue a follow-up query for rich data (ADR-028).
type MutationResult struct {
	ID     string
	Status string
}
