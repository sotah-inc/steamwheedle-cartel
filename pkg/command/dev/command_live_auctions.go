package dev

import (
	"os"
	"os/signal"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	devState "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/dev"
)

func LiveAuctions(config devState.LiveAuctionsStateConfig) error {
	logging.Info("Starting live-auctions")

	// establishing a state
	laState, err := devState.NewLiveAuctionsState(config)
	if err != nil {
		return err
	}

	// opening all listeners
	if err := laState.Listeners.Listen(); err != nil {
		return err
	}

	// starting up a collector
	collectorStop := make(sotah.WorkerStopChan)
	onCollectorStop := make(sotah.WorkerStopChan)
	onCollectorStop = laState.StartCollector(collectorStop)

	// catching SIGINT
	logging.Info("Waiting for SIGINT")
	sigIn := make(chan os.Signal, 1)
	signal.Notify(sigIn, os.Interrupt)
	<-sigIn

	logging.Info("Caught SIGINT, exiting")

	// stopping listeners
	laState.Listeners.Stop()

	// stopping collector
	logging.Info("Stopping collector")
	collectorStop <- struct{}{}

	logging.Info("Waiting for collector to stop")
	<-onCollectorStop

	logging.Info("Exiting")
	return nil
}
