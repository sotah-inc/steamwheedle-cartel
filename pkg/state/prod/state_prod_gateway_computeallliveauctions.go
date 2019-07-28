package prod

import (
	"github.com/sotah-inc/steamwheedle-cartel/pkg/act"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
)

func (sta GatewayState) RunComputeAllLiveAuctions(tuples sotah.RegionRealmTimestampTuples) error {
	// generating an act client
	logging.WithField("endpoint-url", sta.actEndpoints.Gateway).Info("Producing act client")
	actClient, err := act.NewClient(sta.actEndpoints.Gateway)
	if err != nil {
		return err
	}

	// calling compute-all-live-auctions on gateway service
	logging.Info("Calling compute-all-live-auctions on gateway service")
	if err := actClient.ComputeAllLiveAuctions(tuples); err != nil {
		return err
	}

	logging.Info("Done calling compute-all-live-auctions")

	return nil
}

func (sta GatewayState) ListenForCallComputeAllLiveAuctions(
	onReady chan interface{},
	stop chan interface{},
	onStopped chan interface{},
) {
	in := make(chan sotah.RegionRealmTimestampTuples)
	go func() {
		for tuples := range in {
			if err := sta.RunComputeAllLiveAuctions(tuples); err != nil {
				logging.WithField("error", err.Error()).Error("Failed to call RunComputeAllLiveAuctions()")

				continue
			}
		}
	}()

	// establishing subscriber config
	config := bus.SubscribeConfig{
		Stop: stop,
		Callback: func(busMsg bus.Message) {
			logging.WithField("bus-msg", busMsg).Info("Received bus-message")

			// parsing the message body
			tuples, err := sotah.NewRegionRealmTimestampTuples(busMsg.Data)
			if err != nil {
				logging.WithField("error", err.Error()).Error("Failed to parse bus message body")

				return
			}

			in <- tuples
		},
		OnReady:   onReady,
		OnStopped: onStopped,
	}

	// starting up worker for the subscription
	go func() {
		if err := sta.IO.BusClient.SubscribeToTopic(string(subjects.CallComputeAllLiveAuctions), config); err != nil {
			logging.WithField("error", err.Error()).Fatal("Failed to subscribe to topic")
		}
	}()
}
