package run

import (
	"log"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
	"github.com/twinj/uuid"
)

type CleanupPubsubTopicsStateConfig struct {
	ProjectId string
}

func NewCleanupPubsubTopicsState(config CleanupPubsubTopicsStateConfig) (CleanupPubsubTopicsState, error) {
	// establishing an initial state
	sta := CleanupPubsubTopicsState{
		State: state.NewState(uuid.NewV4(), true),
	}

	var err error

	sta.IO.BusClient, err = bus.NewClient(config.ProjectId, "run-cleanup-pubsub-topics")
	if err != nil {
		log.Fatalf("Failed to create new store client: %s", err.Error())

		return CleanupPubsubTopicsState{}, err
	}

	return sta, nil
}

type CleanupPubsubTopicsState struct {
	state.State
}
