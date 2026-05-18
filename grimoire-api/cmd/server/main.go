package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/chwarner-solo/grimoire/grimoire-api/internal/generated"
	"github.com/chwarner-solo/grimoire/grimoire-api/internal/middleware"
	"github.com/chwarner-solo/grimoire/grimoire-api/internal/resolver"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	res := resolver.NewResolver(buildConfig())

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: res,
	}))

	mux := http.NewServeMux()
	mux.Handle("/query", middleware.AuthMiddleware(srv))
	mux.Handle("/", playground.Handler("Grimoire API", "/query"))

	log.Printf("grimoire-api listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// buildConfig wires all interactors. Infrastructure adapters are stubbed with
// nil until Firestore/Firebase implementations are provided.
func buildConfig() resolver.ResolverConfig {
	// TODO: replace nil values with real infrastructure adapters
	// (FirestoreGameRepo, FirestoreEventBus, FirebaseAuthAdapter, etc.)
	return resolver.ResolverConfig{
		Auth: nil, // TODO: sharedinfra.NewFirebaseCallerIdentityAdapter(firebaseApp)

		CreateGame:  nil, // TODO: gameinteractor.NewCreateGameInteractor(gameRepo, bus)
		ArchiveGame: nil,

		CreateCampaign:         nil,
		AddCharacterToCampaign: nil,
		BeginCampaignFormation: nil,
		StartFirstSession:      nil,
		StartNewSession:        nil,
		EndSession:             nil,
		MoveParty:              nil,
		CompleteCampaign:       nil,

		CreateMasterBeat:           nil,
		UpdateBeatContent:          nil,
		AddBeatPrerequisite:        nil,
		AddActToMasterNarrative:    nil,
		AddSecretToMasterNarrative: nil,
		AddLoreToMasterNarrative:   nil,
		DiscoverBeat:               nil,
		CreateCampaignBeat:         nil,
		PromoteBeatToMaster:        nil,

		CreateFaction:      nil,
		BeginFactionDraft:  nil,
		ActivateFaction:    nil,
		AddFactionMember:   nil,
		AddStandingLevel:   nil,
		DeclareAlly:        nil,
		DeclareWar:         nil,
		MarkFactionDormant: nil,
		ReactivateFaction:  nil,
		ArchiveFaction:     nil,

		CreateLocation:   nil,
		ActivateLocation: nil,
		AddScene:         nil,
		ConnectLocations: nil,
		ArchiveLocation:  nil,

		CreateNPC:                    nil,
		BeginNPCDraft:                nil,
		UpdateNPCContent:             nil,
		ActivateNPC:                  nil,
		ArchiveNPC:                   nil,
		CreatePlayerCharacter:        nil,
		UpdatePlayerCharacterContent: nil,
		RetirePlayerCharacter:        nil,
		CreateMacGuffin:              nil,
		UpdateMacGuffinContent:       nil,
		AssignMacGuffinToNPC:         nil,
		AssignMacGuffinToPC:          nil,
		DestroyMacGuffin:             nil,

		RevealEntity: nil,
	}
}
