package value

// LocationType classifies the kind of location.
type LocationType string

const (
	LocationTypeWorld      LocationType = "world"
	LocationTypeRegion     LocationType = "region"
	LocationTypeSettlement LocationType = "settlement"
	LocationTypeBuilding   LocationType = "building"
	LocationTypeDungeon    LocationType = "dungeon"
	LocationTypeWilderness LocationType = "wilderness"
	LocationTypePlane      LocationType = "plane"
)
