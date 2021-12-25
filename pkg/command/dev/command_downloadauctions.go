package dev

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	devState "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/dev"
)

func DownloadAuctions(config devState.DownloadAuctionsStateConfig) error {
	logging.Info("starting DownloadAuctions()")

	startTime := time.Now()

	// establishing a state
	sta, err := devState.NewDownloadAuctionsState(config)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to establish api-state")

		return err
	}

	if err := sta.Run(); err != nil {
		return err
	}

	logging.WithField(
		"duration-in-ms",
		time.Since(startTime).Milliseconds(),
	).Info("finished DownloadAuctions()")

	return nil
}
