package dev

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (laState LiveAuctionsState) StartCollector(stopChan sotah.WorkerStopChan) sotah.WorkerStopChan {
	laState.collectStats()

	onStop := make(sotah.WorkerStopChan)
	go func() {
		ticker := time.NewTicker(30 * time.Minute)

		logging.Info("Starting collector")
	outer:
		for {
			select {
			case <-ticker.C:
				laState.collectStats()
			case <-stopChan:
				ticker.Stop()

				break outer
			}
		}

		onStop <- struct{}{}
	}()

	return onStop
}
