package prod

import (
	"os"
	"os/signal"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	prodState "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/prod"
)

func ProdMetrics(config prodState.ProdMetricsStateConfig) error {
	logging.Info("Starting prod-metrics")

	// establishing a state
	metricsState, err := prodState.NewProdMetricsState(config)
	if err != nil {
		logging.WithField("error", err.Error()).Error("Failed to establish prod-metrics state")

		return err
	}

	// opening all listeners
	if err := metricsState.Listeners.Listen(); err != nil {
		return err
	}

	// opening all bus-listeners
	logging.Info("Opening all bus-listeners")
	metricsState.BusListeners.Listen()

	// catching SIGINT
	logging.Info("Waiting for SIGINT")
	sigIn := make(chan os.Signal, 1)
	signal.Notify(sigIn, os.Interrupt)
	<-sigIn

	logging.Info("Caught SIGINT, exiting")

	// stopping listeners
	metricsState.Listeners.Stop()

	logging.Info("Exiting")
	return nil
}
