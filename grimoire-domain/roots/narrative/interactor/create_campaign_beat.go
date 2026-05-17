package interactor

import (
	"context"
	"fmt"
	"time"

	narrativeentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/narrative/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

type CreateCampaignBeatRequest struct {
	CallerID   string
	BeatID     identity.BeatID
	CampaignID identity.CampaignID
	GameID     identity.GameID
	Name       string
	Source     event.Source
}

type CreateCampaignBeatResult struct {
	Beat   *narrativeentity.Beat
	Events []event.Event
}

type CreateCampaignBeatInteractor struct {
	gameRepo GameRepository
	cnRepo   CampaignNarrativeRepository
	beatRepo BeatRepository
	bus      event.EventBus
}

func NewCreateCampaignBeatInteractor(gameRepo GameRepository, cnRepo CampaignNarrativeRepository, beatRepo BeatRepository, bus event.EventBus) *CreateCampaignBeatInteractor {
	return &CreateCampaignBeatInteractor{gameRepo: gameRepo, cnRepo: cnRepo, beatRepo: beatRepo, bus: bus}
}

func (i *CreateCampaignBeatInteractor) Execute(ctx context.Context, req CreateCampaignBeatRequest) (CreateCampaignBeatResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return CreateCampaignBeatResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return CreateCampaignBeatResult{}, shared.ErrUnauthorized
	}

	cn, err := i.cnRepo.FindByCampaign(ctx, req.CampaignID)
	if err != nil {
		return CreateCampaignBeatResult{}, fmt.Errorf("%w: %v", ErrCampaignNarrativeNotFound, err)
	}

	beat, events, err := narrativeentity.CreateCampaignBeat(req.BeatID, req.Name, req.GameID, req.CampaignID, req.Source)
	if err != nil {
		return CreateCampaignBeatResult{}, err
	}

	if err := i.beatRepo.Save(ctx, beat); err != nil {
		return CreateCampaignBeatResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	cn.AddCampaignBeat(beat.BeatID())
	if err := i.cnRepo.Save(ctx, cn); err != nil {
		return CreateCampaignBeatResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	for _, evt := range events {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.BeatID.GrimoireID),
			AggregateType: event.AggregateBeat,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return CreateCampaignBeatResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return CreateCampaignBeatResult{Beat: beat, Events: events}, nil
}
