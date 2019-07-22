package prod

import (
	"os"
	"os/signal"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	prodState "github.com/sotah-inc/steamwheedle-cartel/pkg/state/prod"
)

func Gateway(config prodState.GatewayStateConfig) error {
	logging.Info("Starting prod-gateway")

	// establishing a state
	sta, err := prodState.NewGatewayState(config)
	if err != nil {
		logging.WithField("error", err.Error()).Error("Failed to establish prod-gateway state")

		return err
	}

	// opening all listeners
	if err := sta.Listeners.Listen(); err != nil {
		return err
	}

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
