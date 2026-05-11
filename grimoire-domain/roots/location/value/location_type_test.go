package value

import "testing"

func TestLocationTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		lt       LocationType
		expected string
	}{
		{"World", LocationTypeWorld, "world"},
		{"Region", LocationTypeRegion, "region"},
		{"Settlement", LocationTypeSettlement, "settlement"},
		{"Building", LocationTypeBuilding, "building"},
		{"Dungeon", LocationTypeDungeon, "dungeon"},
		{"Wilderness", LocationTypeWilderness, "wilderness"},
		{"Plane", LocationTypePlane, "plane"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.lt) != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, string(tt.lt))
			}
		})
	}
}
