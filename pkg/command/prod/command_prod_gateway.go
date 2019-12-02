package prod

import (
	"os"
	"os/signal"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	prodState "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/prod"
)

func Gateway(config prodState.GatewayStateConfig) error {
	logging.Info("Starting prod-gateway")

	// establishing a state
	sta, err := prodState.NewGatewayState(config)
	if err != nil {
		logging.WithField("error", err.Error()).Error("Failed to establish prod-gateway state")

		return err
	}

	// opening all bus-listeners
	sta.BusListeners.Listen()

	// catching SIGINT
	logging.Info("Waiting for SIGINT")
	sigIn := make(chan os.Signal, 1)
	signal.Notify(sigIn, os.Interrupt)
	<-sigIn

	logging.Info("Caught SIGINT, exiting")

	// stopping listeners
	sta.Listeners.Stop()

	logging.Info("Exiting")

	return nil
}
