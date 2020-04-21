package dev

import (
	"os"
	"os/signal"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	devState "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/dev"
)

func Api(config devState.ApiStateConfig) error {
	logging.Info("starting api")

	// establishing a state
	apiState, err := devState.NewAPIState(config)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to establish api-state")

		return err
	}

	// opening all listeners
	if err := apiState.Listeners.Listen(); err != nil {
		return err
	}

	// starting up a collector
	collectorStop := make(sotah.WorkerStopChan)
	onCollectorStop := apiState.StartCollector(collectorStop)

	// catching SIGINT
	logging.Info("waiting for SIGINT")
	sigIn := make(chan os.Signal, 1)
	signal.Notify(sigIn, os.Interrupt)
	<-sigIn

	logging.Info("caught SIGINT, exiting")

	// stopping listeners
	apiState.Listeners.Stop()

	logging.Info("stopping collector")
	collectorStop <- struct{}{}

	logging.Info("waiting for collector to stop")
	<-onCollectorStop

	logging.Info("exiting")

	return nil
}
