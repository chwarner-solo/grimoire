package event

// Source identifies where an event originated.
type Source string

const (
	SourceObsidian Source = "obsidian"
	SourceFoundry  Source = "foundry"
	SourceGrimoire Source = "grimoire"
)
