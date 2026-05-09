package entity

import "strings"

// Start transitions a NewSession to InProgressSession.
func (s *newSession) Start() (InProgressSession, error) {
	return &inProgressSession{s.sessionCore}, nil
}

// End transitions an InProgressSession to CompletedSession.
func (s *inProgressSession) End() (CompletedSession, error) {
	return &completedSession{s.sessionCore}, nil
}

// Summarize transitions a CompletedSession to SummarizedSession.
// Notes must be non-empty.
func (s *completedSession) Summarize(notes string) (SummarizedSession, error) {
	if strings.TrimSpace(notes) == "" {
		return nil, ErrNotesRequired
	}
	core := s.sessionCore
	core.notes = notes
	return &summarizedSession{core}, nil
}

// Close transitions a SummarizedSession to IdleSession (terminal).
func (s *summarizedSession) Close() (IdleSession, error) {
	return &idleSession{s.sessionCore}, nil
}
