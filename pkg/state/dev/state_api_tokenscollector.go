package dev

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/metric"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/resolver"
)

func (sta *APIState) collectRegionTokens() {
	logging.Info("Collecting region-tokens")

	// going over the list of regions
	startTime := time.Now()

	result := database.RegionTokenHistory{}
	for job := range sta.IO.Resolver.GetTokens(resolver.NewRegionHostnameTuples(sta.Regions)) {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("Failed to fetch token for region")

			continue
		}

		result[job.Tuple.RegionName] = database.TokenHistory{job.Info.LastUpdatedTimestamp: job.Info.Price}
	}

	if err := sta.IO.Databases.TokensDatabase.PersistHistory(result); err != nil {
		logging.WithField("error", err.Error()).Error("Failed to persist region token-histories")

		return
	}

	duration := time.Since(startTime)
	sta.IO.Reporter.Report(metric.Metrics{
		"tokenscollector_intake_duration": int(duration) / 1000 / 1000 / 1000,
	})
	logging.Info("Finished tokens-collector")
}
