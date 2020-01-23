package dev

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

func (laState LiveAuctionsState) collectStats() {
	logging.Info("Collecting stats")

	for job := range laState.IO.Databases.LiveAuctionsDatabases.PersistRealmStats(laState.Statuses) {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("Failed to persist realm stats")
		}
	}

	logging.Info("Finished stats-collector")
}
