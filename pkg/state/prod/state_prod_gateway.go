package prod

import (
	"github.com/sotah-inc/steamwheedle-cartel/pkg/messenger"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
	"github.com/twinj/uuid"
)

type GatewayStateConfig struct {
	GCloudProjectID string

	MessengerHost string
	MessengerPort int
}

func NewGatewayState(config GatewayStateConfig) (GatewayState, error) {
	// establishing an initial state
	sta := GatewayState{
		State: state.NewState(uuid.NewV4(), true),
	}

	// connecting to the messenger host
	mess, err := messenger.NewMessenger(config.MessengerHost, config.MessengerPort)
	if err != nil {
		return GatewayState{}, err
	}
	sta.IO.Messenger = mess

	// establishing bus-listeners
	sta.BusListeners = state.NewBusListeners(state.SubjectBusListeners{
		subjects.CallDownloadAllAuctions: sta.ListenForCallDownloadAllAuctions,
	})

	return sta, nil
}

type GatewayState struct {
	state.State
}
