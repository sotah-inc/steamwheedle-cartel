package prod

import (
	"github.com/sotah-inc/steamwheedle-cartel/pkg/act"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
)

func (sta GatewayState) RunCleanupExpiredManifests() error {
	// generating an act client
	logging.WithField("endpoint-url", sta.actEndpoints.CleanupExpiredManifests).Info("Producing act client")
	actClient, err := act.NewClient(sta.actEndpoints.Gateway)
	if err != nil {
		return err
	}

	// calling cleanup-all-expired-manifests on gateway service
	logging.Info("Calling cleanup-all-expired-manifests on gateway service")
	if err := actClient.CleanupAllExpiredManifests(); err != nil {
		return err
	}

	logging.Info("Done calling cleanup-all-expired-manifests")

	return nil
}

func (sta GatewayState) ListenForCallCleanupExpiredManifests(
	onReady chan interface{},
	stop chan interface{},
	onStopped chan interface{},
) {
	in := make(chan interface{})
	go func() {
		for range in {
			if err := sta.RunCleanupExpiredManifests(); err != nil {
				logging.WithField("error", err.Error()).Error("Failed to call RunCleanupExpiredManifests()")

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
		if err := sta.IO.BusClient.SubscribeToTopic(string(subjects.CallCleanupAllExpiredManifests), config); err != nil {
			logging.WithField("error", err.Error()).Fatal("Failed to subscribe to topic")
		}
	}()
}
