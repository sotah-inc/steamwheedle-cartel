package prod

import (
	"os"
	"os/signal"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	prodState "github.com/sotah-inc/steamwheedle-cartel/pkg/state/prod"
)

func PubsubTopicsMonitor(config prodState.PubsubTopicsMonitorStateConfig) error {
	logging.Info("Starting prod-gateway")

	// establishing a state
	sta, err := prodState.NewPubsubTopicsMonitorState(config)
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
	sta.BusListeners.Stop()

	logging.Info("Exiting")

	return nil
}
