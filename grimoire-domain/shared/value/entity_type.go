package value

// EntityType identifies the kind of domain entity being acted upon.
type EntityType string

const (
	EntityTypeNPC             EntityType = "npc"
	EntityTypePlayerCharacter EntityType = "player_character"
	EntityTypeMacGuffin       EntityType = "macguffin"
	EntityTypeFaction         EntityType = "faction"
)
