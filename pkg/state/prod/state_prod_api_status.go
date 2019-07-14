package prod

import (
	"encoding/json"
	"errors"

	nats "github.com/nats-io/go-nats"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	bCodes "github.com/sotah-inc/steamwheedle-cartel/pkg/bus/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/messenger"
	mCodes "github.com/sotah-inc/steamwheedle-cartel/pkg/messenger/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
)

func (apiState ApiState) ListenForBusStatus(
	onReady chan interface{},
	stop chan interface{},
	onStopped chan interface{},
) {
	// establishing subscriber config
	config := bus.SubscribeConfig{
		Stop: stop,
		Callback: func(busMsg bus.Message) {
			reply := bus.NewMessage()

			sr, err := state.NewStatusRequest([]byte(busMsg.Data))
			if err != nil {
				reply.Err = err.Error()
				reply.Code = bCodes.MsgJSONParseError
				if _, err := apiState.IO.BusClient.ReplyTo(busMsg, reply); err != nil {
					logging.WithField("error", err.Error()).Error("Failed to reply")

					return
				}

				return
			}

			reg, err := func() (sotah.Region, error) {
				for _, r := range apiState.Regions {
					if r.Name == sr.RegionName {
						return r, nil
					}
				}

				return sotah.Region{}, errors.New("could not find region")
			}()
			if err != nil {
				reply.Err = err.Error()
				reply.Code = bCodes.NotFound
				if _, err := apiState.IO.BusClient.ReplyTo(busMsg, reply); err != nil {
					logging.WithField("error", err.Error()).Error("Failed to reply")

					return
				}

				return
			}

			regionStatus, ok := apiState.Statuses[reg.Name]
			if !ok {
				reply.Err = "Region found but not in Statuses"
				reply.Code = bCodes.NotFound
				if _, err := apiState.IO.BusClient.ReplyTo(busMsg, reply); err != nil {
					logging.WithField("error", err.Error()).Error("Failed to reply")

					return
				}

				return
			}

			encodedStatus, err := json.Marshal(regionStatus)
			if err != nil {
				reply.Err = err.Error()
				reply.Code = bCodes.GenericError
				if _, err := apiState.IO.BusClient.ReplyTo(busMsg, reply); err != nil {
					logging.WithField("error", err.Error()).Error("Failed to reply")

					return
				}

				return
			}

			reply.Data = string(encodedStatus)
			if _, err := apiState.IO.BusClient.ReplyTo(busMsg, reply); err != nil {
				logging.WithField("error", err.Error()).Error("Failed to reply")

				return
			}
		},
		OnReady:   onReady,
		OnStopped: onStopped,
	}

	// starting up worker for the subscription
	go func() {
		if err := apiState.IO.BusClient.SubscribeToTopic(string(subjects.Status), config); err != nil {
			logging.WithField("error", err.Error()).Fatal("Failed to subscribe to topic")
		}
	}()
}

func (apiState ApiState) ListenForMessengerStatus(stop state.ListenStopChan) error {
	err := apiState.IO.Messenger.Subscribe(string(subjects.Status), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		sr, err := state.NewStatusRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.MsgJSONParseError
			apiState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		reg, err := func() (sotah.Region, error) {
			for _, r := range apiState.Regions {
				if r.Name == sr.RegionName {
					return r, nil
				}
			}

			return sotah.Region{}, errors.New("could not find region")
		}()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.NotFound
			apiState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		regionStatus, ok := apiState.Statuses[reg.Name]
		if !ok {
			m.Err = "Region found but not in Statuses"
			m.Code = mCodes.NotFound
			apiState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		encodedStatus, err := json.Marshal(regionStatus)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			apiState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = string(encodedStatus)
		apiState.IO.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
