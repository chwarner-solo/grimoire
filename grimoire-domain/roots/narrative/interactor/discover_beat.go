package interactor

import (
	"context"
	"fmt"
	"time"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

type DiscoverBeatRequest struct {
	CallerID            string
	CampaignNarrativeID identity.CampaignNarrativeID
	BeatID              identity.BeatID
	GameID              identity.GameID
	Source              event.Source
}

type DiscoverBeatResult struct {
	Events []event.Event
}

type DiscoverBeatInteractor struct {
	gameRepo GameRepository
	cnRepo   CampaignNarrativeRepository
	beatRepo BeatRepository
	bus      event.EventBus
}

func NewDiscoverBeatInteractor(gameRepo GameRepository, cnRepo CampaignNarrativeRepository, beatRepo BeatRepository, bus event.EventBus) *DiscoverBeatInteractor {
	return &DiscoverBeatInteractor{gameRepo: gameRepo, cnRepo: cnRepo, beatRepo: beatRepo, bus: bus}
}

func (i *DiscoverBeatInteractor) Execute(ctx context.Context, req DiscoverBeatRequest) (DiscoverBeatResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return DiscoverBeatResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return DiscoverBeatResult{}, shared.ErrUnauthorized
	}

	cn, err := i.cnRepo.Load(ctx, req.CampaignNarrativeID)
	if err != nil {
		return DiscoverBeatResult{}, fmt.Errorf("%w: %v", ErrCampaignNarrativeNotFound, err)
	}

	beat, err := i.beatRepo.Load(ctx, req.BeatID)
	if err != nil {
		return DiscoverBeatResult{}, fmt.Errorf("%w: %v", ErrBeatNotFound, err)
	}

	events, err := cn.DiscoverBeat(beat)
	if err != nil {
		return DiscoverBeatResult{}, err
	}

	if err := i.cnRepo.Save(ctx, cn); err != nil {
		return DiscoverBeatResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	for _, evt := range events {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.CampaignNarrativeID.GrimoireID),
			AggregateType: event.AggregateCampaignNarrative,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return DiscoverBeatResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return DiscoverBeatResult{Events: events}, nil
}
