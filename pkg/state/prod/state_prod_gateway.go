package prod

import (
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/hell"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
	"github.com/twinj/uuid"
)

type GatewayStateConfig struct {
	ProjectId string
}

func NewGatewayState(config GatewayStateConfig) (GatewayState, error) {
	// establishing an initial state
	sta := GatewayState{
		State: state.NewState(uuid.NewV4(), true),
	}

	var err error

	// connecting to hell
	sta.IO.HellClient, err = hell.NewClient(config.ProjectId)
	if err != nil {
		logging.WithField("error", err.Error()).Error("Failed to connect to hell")

		return GatewayState{}, err
	}

	sta.actEndpoints, err = sta.IO.HellClient.GetActEndpoints()
	if err != nil {
		logging.WithField("error", err.Error()).Error("Failed to fetch act endpoints")

		return GatewayState{}, err
	}

	// establishing a bus
	logging.Info("Connecting bus-client")
	sta.IO.BusClient, err = bus.NewClient(config.ProjectId, "prod-gateway")
	if err != nil {
		return GatewayState{}, err
	}

	// establishing bus-listeners
	sta.BusListeners = state.NewBusListeners(state.SubjectBusListeners{
		subjects.CallDownloadAllAuctions:    sta.ListenForCallDownloadAllAuctions,
		subjects.CallCleanupAllManifests:    sta.ListenForCallCleanupAllManifests,
		subjects.CallCleanupAllAuctions:     sta.ListenForCallCleanupAllAuctions,
		subjects.CallComputeAllLiveAuctions: sta.ListenForCallComputeAllLiveAuctions,
	})

	return sta, nil
}

type GatewayState struct {
	state.State

	actEndpoints hell.ActEndpoints
}
