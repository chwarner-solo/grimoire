package event

// AggregateType identifies which aggregate root an event belongs to.
type AggregateType string

const (
	AggregateGame              AggregateType = "game"
	AggregateCampaign          AggregateType = "campaign"
	AggregateSession           AggregateType = "session"
	AggregateCharacter         AggregateType = "character"
	AggregateLocation          AggregateType = "location"
	AggregateMasterNarrative   AggregateType = "master_narrative"
	AggregateCampaignNarrative AggregateType = "campaign_narrative"
	AggregateBeat              AggregateType = "beat"
	AggregateFaction           AggregateType = "faction"
)
