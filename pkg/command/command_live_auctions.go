package command

import (
	"os"
	"os/signal"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
)

func LiveAuctions(config state.LiveAuctionsStateConfig) error {
	logging.Info("Starting live-auctions")

	// establishing a state
	laState, err := state.NewLiveAuctionsState(config)
	if err != nil {
		return err
	}

	// opening all listeners
	if err := laState.Listeners.Listen(); err != nil {
		return err
	}

	// catching SIGINT
	logging.Info("Waiting for SIGINT")
	sigIn := make(chan os.Signal, 1)
	signal.Notify(sigIn, os.Interrupt)
	<-sigIn

	logging.Info("Caught SIGINT, exiting")

	// stopping listeners
	laState.Listeners.Stop()

	logging.Info("Exiting")
	return nil
}
