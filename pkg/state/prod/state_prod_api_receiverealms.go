package prod

import (
	"encoding/json"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	bCodes "github.com/sotah-inc/steamwheedle-cartel/pkg/bus/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah/gameversions"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
)

func (apiState ApiState) ListenForReceiveRealms(
	onReady chan interface{},
	stop chan interface{},
	onStopped chan interface{},
) {
	// establishing subscriber config
	config := bus.SubscribeConfig{
		Stop: stop,
		Callback: func(busMsg bus.Message) {
			m := bus.NewMessage()

			// parsing bus-message request body
			var regionRealmSlugs map[blizzard.RegionName][]blizzard.RealmSlug
			if err := json.Unmarshal([]byte(busMsg.Data), &regionRealmSlugs); err != nil {
				m.Err = err.Error()
				m.Code = bCodes.GenericError
				if _, err := apiState.IO.BusClient.ReplyTo(busMsg, m); err != nil {
					logging.WithField("error", err.Error()).Error("Failed to reply to bus message")

					return
				}

				return
			}

			hellRegionRealms, err := apiState.IO.HellClient.GetRegionRealms(regionRealmSlugs, gameversions.Retail)
			if err != nil {
				m.Err = err.Error()
				m.Code = bCodes.GenericError
				if _, err := apiState.IO.BusClient.ReplyTo(busMsg, m); err != nil {
					logging.WithField("error", err.Error()).Error("Failed to reply to bus message")

					return
				}

				return
			}

			apiState.HellRegionRealms = apiState.HellRegionRealms.Merge(hellRegionRealms)

			if _, err := apiState.IO.BusClient.ReplyTo(busMsg, m); err != nil {
				logging.WithField("error", err.Error()).Error("Failed to reply to bus message")

				return
			}
		},
		OnReady:   onReady,
		OnStopped: onStopped,
	}

	// starting up worker for the subscription
	go func() {
		if err := apiState.IO.BusClient.SubscribeToTopic(string(subjects.ReceiveRealms), config); err != nil {
			logging.WithField("error", err.Error()).Fatal("Failed to subscribe to topic")
		}
	}()
}
