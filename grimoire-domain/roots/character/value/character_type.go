package value

// CharacterType distinguishes NPC from PlayerCharacter across bounded contexts.
type CharacterType string

const (
	CharacterTypeNPC             CharacterType = "npc"
	CharacterTypePlayerCharacter CharacterType = "player-character"
)
