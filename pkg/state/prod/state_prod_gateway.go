package prod

import (
	nats "github.com/nats-io/go-nats"
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

	// establishing messenger-listeners
	sta.Listeners = state.NewListeners(state.SubjectListeners{
		subjects.CallGateway: sta.ListenForCallGateway,
	})

	return sta, nil
}

type GatewayState struct {
	state.State
}

func (sta GatewayState) ListenForCallGateway(stop state.ListenStopChan) error {
	err := sta.IO.Messenger.Subscribe(string(subjects.Items), stop, func(natsMsg nats.Msg) {
		sta.IO.Messenger.ReplyTo(natsMsg, messenger.NewMessage())
	})
	if err != nil {
		return err
	}

	return nil
}
