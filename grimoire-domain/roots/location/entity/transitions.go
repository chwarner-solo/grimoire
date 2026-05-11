package entity

import (
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- NewLocation transitions ---

func (l *newLocation) BeginDraft(source event.Source) (DraftLocation, []event.Event, error) {
	evt := event.EntityUpdated{
		EntityID: l.id.String(),
		Field:    "status",
		OldValue: "new",
		NewValue: "draft",
		Source:   source,
	}
	return &draftLocation{locationCore: l.locationCore}, []event.Event{evt}, nil
}

// --- DraftLocation transitions ---

func (l *draftLocation) AddScene(sceneID identity.SceneID, source event.Source) (DraftLocation, []event.Event, error) {
	if sceneID.IsZero() {
		return nil, nil, ErrSceneIDRequired
	}
	if l.hasScene(sceneID) {
		return nil, nil, ErrSceneAlreadyExists
	}
	l.sceneIDs = append(l.sceneIDs, sceneID)
	evt := event.EntityLinked{
		EntityAID:    l.id.String(),
		EntityBID:    sceneID.String(),
		Relationship: "has_scene",
		Source:       source,
	}
	return l, []event.Event{evt}, nil
}

func (l *draftLocation) AddChild(childID identity.LocationID, source event.Source) (DraftLocation, []event.Event, error) {
	if childID.IsZero() {
		return nil, nil, ErrLocationIDRequired
	}
	if l.hasChild(childID) {
		return nil, nil, ErrChildAlreadyExists
	}
	l.childLocationIDs = append(l.childLocationIDs, childID)
	evt := event.EntityLinked{
		EntityAID:    l.id.String(),
		EntityBID:    childID.String(),
		Relationship: "contains",
		Source:       source,
	}
	return l, []event.Event{evt}, nil
}

func (l *draftLocation) Activate(source event.Source) (ActiveLocation, []event.Event, error) {
	if len(l.sceneIDs) == 0 {
		return nil, nil, ErrCannotActivateWithoutScenes
	}
	evt := event.EntityUpdated{
		EntityID: l.id.String(),
		Field:    "status",
		OldValue: "draft",
		NewValue: "active",
		Source:   source,
	}
	return &activeLocation{locationCore: l.locationCore}, []event.Event{evt}, nil
}

// --- ActiveLocation transitions ---

func (l *activeLocation) AddScene(sceneID identity.SceneID, source event.Source) (ActiveLocation, []event.Event, error) {
	if sceneID.IsZero() {
		return nil, nil, ErrSceneIDRequired
	}
	if l.hasScene(sceneID) {
		return nil, nil, ErrSceneAlreadyExists
	}
	l.sceneIDs = append(l.sceneIDs, sceneID)
	evt := event.EntityLinked{
		EntityAID:    l.id.String(),
		EntityBID:    sceneID.String(),
		Relationship: "has_scene",
		Source:       source,
	}
	return l, []event.Event{evt}, nil
}

func (l *activeLocation) AddChild(childID identity.LocationID, source event.Source) (ActiveLocation, []event.Event, error) {
	if childID.IsZero() {
		return nil, nil, ErrLocationIDRequired
	}
	if l.hasChild(childID) {
		return nil, nil, ErrChildAlreadyExists
	}
	l.childLocationIDs = append(l.childLocationIDs, childID)
	evt := event.EntityLinked{
		EntityAID:    l.id.String(),
		EntityBID:    childID.String(),
		Relationship: "contains",
		Source:       source,
	}
	return l, []event.Event{evt}, nil
}

func (l *activeLocation) ConnectTo(locationID identity.LocationID, source event.Source) (ActiveLocation, []event.Event, error) {
	if locationID.String() == l.id.String() {
		return nil, nil, ErrCannotConnectToSelf
	}
	if l.isConnectedTo(locationID) {
		return nil, nil, ErrAlreadyConnectedTo{LocationID: locationID}
	}
	l.connectedLocationIDs = append(l.connectedLocationIDs, locationID)
	evt := event.EntityLinked{
		EntityAID:    l.id.String(),
		EntityBID:    locationID.String(),
		Relationship: "connects_to",
		Source:       source,
	}
	return l, []event.Event{evt}, nil
}

func (l *activeLocation) GoIdle(source event.Source) (IdleLocation, []event.Event, error) {
	evt := event.EntityUpdated{
		EntityID: l.id.String(),
		Field:    "status",
		OldValue: "active",
		NewValue: "idle",
		Source:   source,
	}
	return &idleLocation{locationCore: l.locationCore}, []event.Event{evt}, nil
}

func (l *activeLocation) Archive(source event.Source) (ArchivedLocation, []event.Event, error) {
	evt := event.EntityUpdated{
		EntityID: l.id.String(),
		Field:    "status",
		OldValue: "active",
		NewValue: "archived",
		Source:   source,
	}
	return &archivedLocation{locationCore: l.locationCore}, []event.Event{evt}, nil
}

// --- IdleLocation transitions ---

func (l *idleLocation) Reactivate(source event.Source) (ActiveLocation, []event.Event, error) {
	evt := event.EntityUpdated{
		EntityID: l.id.String(),
		Field:    "status",
		OldValue: "idle",
		NewValue: "active",
		Source:   source,
	}
	return &activeLocation{locationCore: l.locationCore}, []event.Event{evt}, nil
}

func (l *idleLocation) Archive(source event.Source) (ArchivedLocation, []event.Event, error) {
	evt := event.EntityUpdated{
		EntityID: l.id.String(),
		Field:    "status",
		OldValue: "idle",
		NewValue: "archived",
		Source:   source,
	}
	return &archivedLocation{locationCore: l.locationCore}, []event.Event{evt}, nil
}
