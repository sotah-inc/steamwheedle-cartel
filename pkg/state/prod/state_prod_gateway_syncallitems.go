package prod

import (
	"github.com/sotah-inc/steamwheedle-cartel/pkg/act"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
)

func (sta GatewayState) RunSyncAllItems(ids blizzard.ItemIds) error {
	// generating an act client
	logging.WithField("endpoint-url", sta.actEndpoints.Gateway).Info("Producing act client for gateway act endpoint")
	actClient, err := act.NewClient(sta.actEndpoints.Gateway)
	if err != nil {
		return err
	}

	// calling sync-all-items on gateway service
	logging.Info("Calling sync-all-items on gateway service")
	if err := actClient.SyncAllItems(ids); err != nil {
		return err
	}

	logging.Info("Done calling sync-all-items")

	return nil
}

func (sta GatewayState) ListenForCallSyncAllItems(
	onReady chan interface{},
	stop chan interface{},
	onStopped chan interface{},
) {
	in := make(chan blizzard.ItemIds)
	go func() {
		for ids := range in {
			if err := sta.RunSyncAllItems(ids); err != nil {
				logging.WithField("error", err.Error()).Error("Failed to call RunSyncAllItems()")

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
			ids, err := blizzard.NewItemIds(busMsg.Data)
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

			in <- ids
		},
		OnReady:   onReady,
		OnStopped: onStopped,
	}

	// starting up worker for the subscription
	go func() {
		if err := sta.IO.BusClient.SubscribeToTopic(string(subjects.CallSyncAllItems), config); err != nil {
			logging.WithField("error", err.Error()).Fatal("Failed to subscribe to topic")
		}
	}()
}
