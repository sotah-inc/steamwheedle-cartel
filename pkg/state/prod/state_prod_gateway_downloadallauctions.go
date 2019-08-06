package prod

import (
	"github.com/sotah-inc/steamwheedle-cartel/pkg/act"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
)

func (sta GatewayState) RunDownloadAllAuctions() error {
	// generating an act client
	logging.WithField("endpoint-url", sta.actEndpoints.Gateway).Info("Producing act client for gateway act endpoint")
	actClient, err := act.NewClient(sta.actEndpoints.Gateway)
	if err != nil {
		return err
	}

	// calling download-all-auctions on gateway service
	logging.Info("Calling download-all-auctions on gateway service")
	if err := actClient.DownloadAllAuctions(); err != nil {
		return err
	}

	logging.Info("Done calling download-all-auctions")

	return nil
}

func (sta GatewayState) ListenForCallDownloadAllAuctions(
	onReady chan interface{},
	stop chan interface{},
	onStopped chan interface{},
) {
	in := make(chan interface{})
	go func() {
		for range in {
			if err := sta.RunDownloadAllAuctions(); err != nil {
				logging.WithField("error", err.Error()).Error("Failed to call RunDownloadAllAuctions()")

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
		if err := sta.IO.BusClient.SubscribeToTopic(string(subjects.CallDownloadAllAuctions), config); err != nil {
			logging.WithField("error", err.Error()).Fatal("Failed to subscribe to topic")
		}
	}()
}
