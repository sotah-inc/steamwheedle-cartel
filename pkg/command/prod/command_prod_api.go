package prod

import (
	"os"
	"os/signal"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	prodState "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/prod"
)

func ProdApi(config prodState.ProdApiStateConfig) error {
	logging.Info("Starting prod-api")

	// establishing a state
	apiState, err := prodState.NewProdApiState(config)
	if err != nil {
		logging.WithField("error", err.Error()).Error("Failed to establish api-state")

		return err
	}

	// opening all listeners
	if err := apiState.Listeners.Listen(); err != nil {
		return err
	}

	// opening all bus-listeners
	logging.Info("Opening all bus-listeners")
	apiState.BusListeners.Listen()

	// catching SIGINT
	logging.Info("Waiting for SIGINT")
	sigIn := make(chan os.Signal, 1)
	signal.Notify(sigIn, os.Interrupt)
	<-sigIn

	logging.Info("Caught SIGINT, exiting")

	// stopping listeners
	apiState.Listeners.Stop()
	apiState.BusListeners.Stop()

	logging.Info("Exiting")
	return nil
}
