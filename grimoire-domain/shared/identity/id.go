package identity

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)


type GrimoireID struct {
	value uuid.UUID
}

func NewGrimoireID() GrimoireID {
	return GrimoireID{value: uuid.New()}
}

func ParseGrimoireID(s string) (GrimoireID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return GrimoireID{}, fmt.Errorf("%w: %s", ErrInvalidID, s)
	}
	return GrimoireID{value: id}, nil
}

func (id GrimoireID) String() string {
	return id.value.String()
}

func (id GrimoireID) IsZero() bool {
	return id.value == uuid.Nil
}

func (id GrimoireID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.value.String())
}

func (id *GrimoireID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := uuid.Parse(s)
	if err != nil {
		return err
	}
	id.value = parsed
	return nil
}
