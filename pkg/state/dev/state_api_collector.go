package dev

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (sta ApiState) Collect() error {
	collectAuctionsResults, err := sta.DiskAuctionsState.CollectAuctions()
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to collect auctions")

		return err
	}

	logging.WithField("item-ids", len(collectAuctionsResults.ItemIds)).Info("found items in auctions")

	if err := sta.ItemsState.CollectItems(collectAuctionsResults.ItemIds); err != nil {
		logging.WithField("error", err.Error()).Error("failed to collect items")

		return err
	}

	if err := sta.TokensState.CollectRegionTokens(sta.RegionState.RegionComposites.ToList()); err != nil {
		logging.WithField("error", err.Error()).Error("failed to collect region-tokens")

		return err
	}

	if err := sta.LiveAuctionsState.LiveAuctionsIntake(collectAuctionsResults.Tuples); err != nil {
		logging.WithField("error", err.Error()).Error("failed to execute live-auctions-intake")

		return err
	}

	return nil
}

func (sta ApiState) StartCollector(stopChan sotah.WorkerStopChan) sotah.WorkerStopChan {
	if err := sta.Collect(); err != nil {
		logging.WithField("error", err.Error()).Error("failed to collect")
	}
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
