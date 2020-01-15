package dev

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (sta *APIState) StartCollector(stopChan sotah.WorkerStopChan) sotah.WorkerStopChan {
	sta.collectRegions()

	onStop := make(sotah.WorkerStopChan)
	go func() {
		ticker := time.NewTicker(20 * time.Minute)

		logging.Info("Starting collector")
	outer:
		for {
			select {
			case <-ticker.C:
				// refreshing the access-token for the Resolver blizz client
				nextClient, err := sta.IO.Resolver.BlizzardClient.RefreshFromHTTP(blizzard.OAuthTokenEndpoint)
				if err != nil {
					logging.WithField("error", err.Error()).Error("Failed to refresh blizzard client")

					continue
				}
				sta.IO.Resolver.BlizzardClient = nextClient

				sta.collectRegions()
			case <-stopChan:
				ticker.Stop()

				break outer
			}
		}

		onStop <- struct{}{}
	}()

	return onStop
}
