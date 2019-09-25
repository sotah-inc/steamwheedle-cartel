package prod

import (
	"os"
	"os/signal"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	prodState "github.com/sotah-inc/steamwheedle-cartel/pkg/state/prod"
)

func Items(config prodState.ItemsStateConfig) error {
	logging.Info("Starting prod-items")

	// establishing a state
	itemsState, err := prodState.NewProdItemsState(config)
	if err != nil {
		logging.WithField("error", err.Error()).Error("Failed to establish prod-items state")

		return err
	}

	// opening all listeners
	if err := itemsState.Listeners.Listen(); err != nil {
		return err
	}

	// opening all bus-listeners
	logging.Info("Opening all bus-listeners")
	itemsState.BusListeners.Listen()

	// catching SIGINT
	logging.Info("Waiting for SIGINT")
	sigIn := make(chan os.Signal, 1)
	signal.Notify(sigIn, os.Interrupt)
	<-sigIn

	logging.Info("Caught SIGINT, exiting")

	// stopping listeners
	itemsState.Listeners.Stop()

	logging.Info("Exiting")
	return nil
}
