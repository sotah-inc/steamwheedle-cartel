package prod

import (
	"github.com/sotah-inc/steamwheedle-cartel/pkg/act"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
)

func (sta GatewayState) RunCleanupAllAuctions() error {
	// generating an act client
	logging.WithField("endpoint-url", sta.actEndpoints.Gateway).Info("Producing act client")
	actClient, err := act.NewClient(sta.actEndpoints.Gateway)
	if err != nil {
		return err
	}

	// calling cleanup-all-auctions on gateway service
	logging.Info("Calling cleanup-all-auctions on gateway service")
	if err := actClient.CleanupAllAuctions(); err != nil {
		return err
	}

	logging.Info("Done calling cleanup-all-auctions")

	return nil
}

func (sta GatewayState) ListenForCallCleanupAllAuctions(
	onReady chan interface{},
	stop chan interface{},
	onStopped chan interface{},
) {
	in := make(chan interface{})
	go func() {
		for range in {
			if err := sta.RunCleanupAllAuctions(); err != nil {
				logging.WithField("error", err.Error()).Error("Failed to call RunCleanupAllAuctions()")

				continue
			}
		}
	}()

	// establishing subscriber config
	config := bus.SubscribeConfig{
		Stop: stop,
		Callback: func(busMsg bus.Message) {
			logging.WithField("bus-msg", busMsg).Info("Received bus-message")
			in <- struct{}{}
		},
		OnReady:   onReady,
		OnStopped: onStopped,
	}

	// starting up worker for the subscription
	go func() {
		if err := sta.IO.BusClient.SubscribeToTopic(string(subjects.CallCleanupAllAuctions), config); err != nil {
			logging.WithField("error", err.Error()).Fatal("Failed to subscribe to topic")
		}
	}()
}
