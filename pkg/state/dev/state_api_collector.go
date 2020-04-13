package dev

import (
	"errors"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (sta ApiState) Collect() error {
	if 1+1 == 2 {
		return errors.New("fail")
	}

	itemIds, err := sta.DiskAuctionsState.CollectAuctions()
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to collect auctions")

		return err
	}

	logging.WithField("item-ids", len(itemIds)).Info("found items in auctions")

	if err := sta.DiskAuctionsState.CollectItems(itemIds); err != nil {
		logging.WithField("error", err.Error()).Error("failed to collect items")

		return err
	}

	return nil
}

func (sta ApiState) StartCollector(stopChan sotah.WorkerStopChan) sotah.WorkerStopChan {
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
