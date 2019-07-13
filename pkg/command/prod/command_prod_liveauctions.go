package prod

import (
	"os"
	"os/signal"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	prodState "github.com/sotah-inc/steamwheedle-cartel/pkg/state/prod"
)

func ProdLiveAuctions(config prodState.ProdLiveAuctionsStateConfig) error {
	logging.Info("Starting prod-liveauctions")

	// establishing a state
	liveAuctionsState, err := prodState.NewProdLiveAuctionsState(config)
	if err != nil {
		logging.WithField("error", err.Error()).Error("Failed to establish prod-liveauctions state")

		return err
	}

	// opening all listeners
	if err := liveAuctionsState.Listeners.Listen(); err != nil {
		return err
	}

	// opening all bus-listeners
	logging.Info("Opening all bus-listeners")
	liveAuctionsState.BusListeners.Listen()

	// catching SIGINT
	logging.Info("Waiting for SIGINT")
	sigIn := make(chan os.Signal, 1)
	signal.Notify(sigIn, os.Interrupt)
	<-sigIn

	logging.Info("Caught SIGINT, exiting")

	// stopping listeners
	liveAuctionsState.Listeners.Stop()

	logging.Info("Exiting")
	return nil
}
