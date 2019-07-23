package prod

import (
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
	"github.com/twinj/uuid"
)

type GatewayStateConfig struct {
	GCloudProjectID string
}

func NewGatewayState(config GatewayStateConfig) (GatewayState, error) {
	// establishing an initial state
	sta := GatewayState{
		State: state.NewState(uuid.NewV4(), true),
	}

	var err error

	// establishing a bus
	logging.Info("Connecting bus-client")
	sta.IO.BusClient, err = bus.NewClient(config.GCloudProjectID, "prod-gateway")
	if err != nil {
		return GatewayState{}, err
	}

	// establishing bus-listeners
	sta.BusListeners = state.NewBusListeners(state.SubjectBusListeners{
		subjects.CallDownloadAllAuctions: sta.ListenForCallDownloadAllAuctions,
	})

	return sta, nil
}

type GatewayState struct {
	state.State
}
