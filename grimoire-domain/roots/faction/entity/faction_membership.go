package entity

import (
	"strings"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// FactionMembership is a child entity representing a character's membership
// in a faction. It has its own ID to enable EntityRevealed targeting.
type FactionMembership struct {
	id            identity.FactionMembershipID
	factionID     identity.FactionID
	characterID   identity.CharacterID
	rank          string
	playerVisible bool
}

// NewFactionMembership constructs a FactionMembership with validation.
// playerVisible defaults to false.
func NewFactionMembership(id identity.FactionMembershipID, factionID identity.FactionID, characterID identity.CharacterID, rank string) (*FactionMembership, error) {
	if id.IsZero() {
		return nil, ErrMembershipIDRequired
	}
	if factionID.IsZero() {
		return nil, ErrFactionIDRequired
	}
	if characterID.IsZero() {
		return nil, ErrCharacterIDRequired
	}
	if strings.TrimSpace(rank) == "" {
		return nil, ErrRankRequired
	}
	return &FactionMembership{
		id:            id,
		factionID:     factionID,
		characterID:   characterID,
		rank:          rank,
		playerVisible: false,
	}, nil
}

func (m *FactionMembership) MembershipID() identity.FactionMembershipID { return m.id }
func (m *FactionMembership) FactionID() identity.FactionID              { return m.factionID }
func (m *FactionMembership) CharacterID() identity.CharacterID          { return m.characterID }
func (m *FactionMembership) Rank() string                               { return m.rank }
func (m *FactionMembership) IsPlayerVisible() bool                      { return m.playerVisible }

// Reveal marks this membership as visible to players.
func (m *FactionMembership) Reveal() {
	m.playerVisible = true
}

// FactionMembershipSnapshot is the DTO used for persistence.
type FactionMembershipSnapshot struct {
	ID            identity.FactionMembershipID
	FactionID     identity.FactionID
	CharacterID   identity.CharacterID
	Rank          string
	PlayerVisible bool
}

func (m *FactionMembership) Snapshot() FactionMembershipSnapshot {
	return FactionMembershipSnapshot{
		ID:            m.id,
		FactionID:     m.factionID,
		CharacterID:   m.characterID,
		Rank:          m.rank,
		PlayerVisible: m.playerVisible,
	}
}

// ReconstituteFactionMembership rebuilds a FactionMembership from a persisted snapshot.
func ReconstituteFactionMembership(snap FactionMembershipSnapshot) *FactionMembership {
	return &FactionMembership{
		id:            snap.ID,
		factionID:     snap.FactionID,
		characterID:   snap.CharacterID,
		rank:          snap.Rank,
		playerVisible: snap.PlayerVisible,
	}
}
