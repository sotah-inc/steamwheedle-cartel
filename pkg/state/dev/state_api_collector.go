package dev

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (sta APIState) Collect() error {
	itemIds, err := sta.DiskAuctionsState.CollectAuctions()
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to collect auctions")

		return err
	}

	logging.WithField("item-ids", len(itemIds)).Info("found items in auctions")

	itemSyncPayload, err := sta.ItemsState.ItemsDatabase.FilterInItemsToSync(itemIds)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to filter in items-to-sync")

		return err
	}

	logging.WithField("item-sync-payload", itemSyncPayload).Info("received item-sync-payload")

	return nil
}

func (sta APIState) StartCollector(stopChan sotah.WorkerStopChan) sotah.WorkerStopChan {
	if err := sta.Collect(); err != nil {
		logging.WithField("error", err.Error()).Error("failed to collect")
	}

	onStop := make(sotah.WorkerStopChan)
	go func() {
		ticker := time.NewTicker(20 * time.Minute)

		logging.Info("starting collector")
	outer:
		for {
			select {
			case <-ticker.C:
				if err := sta.BlizzardState.BlizzardClient.RefreshFromHTTP(blizzardv2.OAuthTokenEndpoint); err != nil {
					logging.WithField("error", err.Error()).Error("failed to refresh blizzard http-client")

					continue
				}

				if err := sta.Collect(); err != nil {
					logging.WithField("error", err.Error()).Error("failed to collect")

					continue
				}
			case <-stopChan:
				ticker.Stop()

				break outer
			}
		}

		onStop <- struct{}{}
	}()

	return onStop
}
