package dev

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	fnState "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/fn"
)

func LoadAreaMaps(config fnState.LoadAreaMapsStateConfig) error {
	logging.Info("Starting api")

	// establishing a state
	sta, err := fnState.NewLoadAreaMapsState(config)
	if err != nil {
		logging.WithField("error", err.Error()).Error("Failed to establish load-area-maps-state")

		return err
	}

	if err := sta.Run(); err != nil {
		logging.WithField("error", err.Error()).Error("Failed to run load-area-maps-state")

		return err
	}

	logging.Info("Exiting")
	return nil
}
