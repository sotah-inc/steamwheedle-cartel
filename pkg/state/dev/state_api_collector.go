package dev

import (
	"errors"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (sta ApiState) Collect() error {
	startTime := time.Now()
	logging.Info("calling ApiState.Collect()")

	if true {
		return errors.New("test")
	}

	if err := sta.Collector.Collect(); err != nil {
		logging.WithField("error", err.Error()).Error("failed to collect")

		return err
	}

	logging.WithField(
		"duration-in-ms",
		time.Since(startTime).Milliseconds(),
	).Info("finished calling ApiState.Collect()")

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
				if err := sta.BlizzardState.BlizzardClient.RefreshFromHTTP(
					blizzardv2.OAuthTokenEndpoint,
				); err != nil {
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
