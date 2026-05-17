package interactor

import (
	"context"
	"fmt"
	"time"

	campaignentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/campaign/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

type CreateCampaignRequest struct {
	CallerID   string
	CampaignID identity.CampaignID
	GameID     identity.GameID
	Name       string
	Source     event.Source
}

type CreateCampaignResult struct {
	Campaign campaignentity.NewCampaign
	Events   []event.Event
}

type CreateCampaignInteractor struct {
	gameRepo     GameRepository
	campaignRepo CampaignRepository
	bus          event.EventBus
}

func NewCreateCampaignInteractor(gameRepo GameRepository, campaignRepo CampaignRepository, bus event.EventBus) *CreateCampaignInteractor {
	return &CreateCampaignInteractor{gameRepo: gameRepo, campaignRepo: campaignRepo, bus: bus}
}

func (i *CreateCampaignInteractor) Execute(ctx context.Context, req CreateCampaignRequest) (CreateCampaignResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return CreateCampaignResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return CreateCampaignResult{}, shared.ErrUnauthorized
	}

	campaign, events, err := campaignentity.CreateCampaign(req.CampaignID, req.Name, req.GameID, req.Source)
	if err != nil {
		return CreateCampaignResult{}, err
	}

	if err := i.campaignRepo.Save(ctx, campaign); err != nil {
		return CreateCampaignResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	toDispatch := append(events, event.EntityLinked{
		EntityAID:    req.GameID.String(),
		EntityBID:    req.CampaignID.String(),
		Relationship: "campaign",
		Source:       req.Source,
	})

	for _, evt := range toDispatch {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.CampaignID.GrimoireID),
			AggregateType: event.AggregateCampaign,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return CreateCampaignResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return CreateCampaignResult{Campaign: campaign, Events: events}, nil
}