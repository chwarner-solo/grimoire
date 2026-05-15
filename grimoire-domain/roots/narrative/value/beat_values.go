package value

// BeatScope indicates whether a beat belongs to the master narrative or a campaign.
type BeatScope string

const (
	BeatScopeMaster    BeatScope = "master"
	BeatScopeCampaign  BeatScope = "campaign"
	BeatScopeCharacter BeatScope = "character"
)

// BeatType categorizes how a beat relates to the narrative structure.
type BeatType string

const (
	BeatTypeRequired         BeatType = "required"
	BeatTypeOptional         BeatType = "optional"
	BeatTypeCampaignSpecific BeatType = "campaign-specific"
	BeatTypeBackground       BeatType = "background"
	BeatTypePersonalArc      BeatType = "personal-arc"
)

// BeatStatus tracks the lifecycle of a beat within the narrative.
type BeatStatus string

const (
	BeatStatusDraft      BeatStatus = "draft"
	BeatStatusActive     BeatStatus = "active"
	BeatStatusDeprecated BeatStatus = "deprecated"
)
