package state

import (
	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/database"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store"
)

func (phState ProdPricelistHistoriesState) Sync() error {
	// gathering region-realms
	regionRealms := map[blizzard.RegionName]sotah.Realms{}
	for regionName, status := range phState.Statuses {
		regionRealms[regionName] = status.Realms
	}

	// gathering existing pricelist-histories versions
	logging.Info("Gathering existing versions")
	versions, err := phState.PricelistHistoriesBase.GetVersions(regionRealms, phState.PricelistHistoriesBucket)
	if err != nil {
		return err
	}

	// trimming matching versions
	logging.Info("Trimming existing versions")
	versionsToSync := sotah.PricelistHistoryVersions{}
	for regionName, realmTimestampVersions := range versions {
		for realmSlug, timestampVersions := range realmTimestampVersions {
			for targetTimestamp, version := range timestampVersions {
				hasBucket, err := phState.IO.Databases.MetaDatabase.HasBucket(regionName, realmSlug)
				if err != nil {
					return err
				}

				if !hasBucket {
					versionsToSync = versionsToSync.Insert(regionName, realmSlug, targetTimestamp, version)

					continue
				}

				hasVersion, err := phState.IO.Databases.MetaDatabase.HasPricelistHistoriesVersion(
					regionName,
					realmSlug,
					targetTimestamp,
				)
				if err != nil {
					return err
				}

				if !hasVersion {
					versionsToSync = versionsToSync.Insert(regionName, realmSlug, targetTimestamp, version)

					continue
				}

				currentVersion, err := phState.IO.Databases.MetaDatabase.GetPricelistHistoriesVersion(
					regionName,
					realmSlug,
					targetTimestamp,
				)
				if err != nil {
					return err
				}

				if currentVersion != version {
					versionsToSync = versionsToSync.Insert(regionName, realmSlug, targetTimestamp, version)

					continue
				}
			}
		}
	}

	// declaring a get-in channel for gathering pricelist-histories
	getInJobs := make(chan store.GetAllPricelistHistoriesInJob)
	getOutJobs := phState.PricelistHistoriesBase.GetAll(getInJobs, phState.PricelistHistoriesBucket)

	// declaring a load-in channel for the pricelist-histories db
	loadInJobs := make(chan database.PricelistHistoryDatabaseEncodedLoadInJob)
	loadOutJobs := phState.IO.Databases.PricelistHistoryDatabases.LoadEncoded(loadInJobs)

	// spinning up a worker for translating get-out-jobs to load-in-jobs
	go func() {
		for outJob := range getOutJobs {
			if outJob.Err != nil {
				logging.WithFields(outJob.ToLogrusFields()).Error("Failed to get pricelist-histories")

				continue
			}

			loadInJobs <- database.PricelistHistoryDatabaseEncodedLoadInJob{
				RegionName:                outJob.RegionName,
				RealmSlug:                 outJob.RealmSlug,
				NormalizedTargetTimestamp: outJob.TargetTimestamp,
				Data:                      outJob.Data,
				VersionId:                 outJob.VersionId,
			}
		}

		close(loadInJobs)
	}()

	// queueing it all up
	go func() {
		jobs := store.NewGetAllPricelistHistoriesInJobs(versionsToSync)
		logging.WithField("jobs", len(jobs)).Info("Queueing up jobs")
		for _, job := range jobs {
			logging.WithFields(logrus.Fields{
				"region":           job.RegionName,
				"realm":            job.RealmSlug,
				"target-timestamp": job.TargetTimestamp,
			}).Info("Loading job")

			getInJobs <- store.GetAllPricelistHistoriesInJob{
				RegionName:      job.RegionName,
				RealmSlug:       job.RealmSlug,
				TargetTimestamp: job.TargetTimestamp,
			}
		}

		close(getInJobs)
	}()

	// waiting for the results to drain out
	versionsToSet := sotah.PricelistHistoryVersions{}
	for job := range loadOutJobs {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("Failed to load job")

			continue
		}

		versionsToSet = versionsToSet.Insert(
			job.RegionName,
			job.RealmSlug,
			job.NormalizedTargetTimestamp,
			job.VersionId,
		)

		logging.WithFields(logrus.Fields{
			"region": job.RegionName,
			"realm":  job.RealmSlug,
		}).Info("Loaded job")
	}

	// setting versions
	if err := phState.IO.Databases.MetaDatabase.SetPricelistHistoriesVersions(versionsToSet); err != nil {
		return err
	}

	return nil
}
