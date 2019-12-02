package prod

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/act"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/bus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/bus/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta GatewayState) RunComputeAllPricelistHistories(tuples sotah.RegionRealmTimestampTuples) error {
	// generating an act client
	logging.WithField("endpoint-url", sta.actEndpoints.Gateway).Info("Producing act client for gateway act endpoint")
	actClient, err := act.NewClient(sta.actEndpoints.Gateway)
	if err != nil {
		return err
	}

	// calling compute-all-pricelist-histories on gateway service
	startTime := time.Now()
	logging.Info("Calling compute-all-pricelist-histories on gateway service")
	if err := actClient.ComputeAllPricelistHistories(tuples); err != nil {
		return err
	}

	logging.WithField(
		"duration",
		int(int64(time.Since(startTime))/1000/1000/1000),
	).Info("Done calling compute-all-pricelist-histories")

	return nil
}

func (sta GatewayState) ListenForCallComputeAllPricelistHistories(
	onReady chan interface{},
	stop chan interface{},
	onStopped chan interface{},
) {
	in := make(chan sotah.RegionRealmTimestampTuples)
	go func() {
		for tuples := range in {
			if err := sta.RunComputeAllPricelistHistories(tuples); err != nil {
				logging.WithField("error", err.Error()).Error("Failed to call RunComputeAllPricelistHistories()")

				continue
			}
		}
	}()

	// establishing subscriber config
	config := bus.SubscribeConfig{
		Stop: stop,
		Callback: func(busMsg bus.Message) {
			logging.WithField("bus-msg-code", busMsg.Code).Info("Received bus-message")

			// parsing the message body
			tuples, err := sotah.NewRegionRealmTimestampTuples(busMsg.Data)
			if err != nil {
				logging.WithField("error", err.Error()).Error("Failed to parse bus message body")

				if err := sta.IO.BusClient.ReplyToWithError(busMsg, err, codes.GenericError); err != nil {
					logging.WithField("error", err.Error()).Error("Failed to reply to message")

					return
				}

				return
			}

			// acking the message
			if _, err := sta.IO.BusClient.ReplyTo(busMsg, bus.NewMessage()); err != nil {
				logging.WithField("error", err.Error()).Error("Failed to reply to message")

				return
			}

			in <- tuples
		},
		OnReady:   onReady,
		OnStopped: onStopped,
	}

	// starting up worker for the subscription
	go func() {
		if err := sta.IO.BusClient.SubscribeToTopic(string(subjects.CallComputeAllPricelistHistories), config); err != nil {
			logging.WithField("error", err.Error()).Fatal("Failed to subscribe to topic")
		}
	}()
}
