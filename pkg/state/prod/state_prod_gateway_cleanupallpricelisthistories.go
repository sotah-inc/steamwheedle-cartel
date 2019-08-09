package prod

import (
	"github.com/sotah-inc/steamwheedle-cartel/pkg/act"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
)

func (sta GatewayState) RunCleanupAllPricelistHistories() error {
	// generating an act client
	logging.WithField("endpoint-url", sta.actEndpoints.Gateway).Info("Producing act client for gateway act endpoint")
	actClient, err := act.NewClient(sta.actEndpoints.Gateway)
	if err != nil {
		return err
	}

	// calling cleanup-all-pricelist-histories on gateway service
	logging.Info("Calling cleanup-all-pricelist-histories on gateway service")
	if err := actClient.CleanupAllPricelistHistories(); err != nil {
		return err
	}

	logging.Info("Done calling cleanup-all-pricelist-histories")

	return nil
}

func (sta GatewayState) ListenForCallCleanupAllPricelistHistories(
	onReady chan interface{},
	stop chan interface{},
	onStopped chan interface{},
) {
	in := make(chan interface{})
	go func() {
		for range in {
			if err := sta.RunCleanupAllPricelistHistories(); err != nil {
				logging.WithField("error", err.Error()).Error("Failed to call RunCleanupAllPricelistHistories()")

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
		if err := sta.IO.BusClient.SubscribeToTopic(string(subjects.CallCleanupAllPricelistHistories), config); err != nil {
			logging.WithField("error", err.Error()).Fatal("Failed to subscribe to topic")
		}
	}()
}
