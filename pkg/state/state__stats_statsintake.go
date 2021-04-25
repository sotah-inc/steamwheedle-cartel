package state

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/base"   // nolint:lll
	StatsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/stats" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta StatsState) ListenForStatsIntake(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.StatsIntake), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		tuples, err := blizzardv2.NewLoadConnectedRealmTuples(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		logging.WithField("tuples", len(tuples)).Info("received")
		if err := sta.StatsIntake(tuples); err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}

func (sta StatsState) StatsIntake(tuples blizzardv2.LoadConnectedRealmTuples) error {
	if err := sta.TuplesIntake(tuples); err != nil {
		return err
	}

	if err := sta.RegionRealmsIntake(sta.Tuples.ToMap()); err != nil {
		return err
	}

	return nil
}

func (sta StatsState) RegionRealmsIntake(
	regionRealmMap map[blizzardv2.RegionName][]blizzardv2.ConnectedRealmId,
) error {
	startTime := time.Now()
	currentTimestamp := sotah.UnixTimestamp(time.Now().Unix())
	retentionLimit := sotah.UnixTimestamp(BaseDatabase.RetentionLimit().Unix())

	for name, ids := range regionRealmMap {
		encodedStats, err := sta.LakeClient.GetEncodedRegionStats(name, ids)
		if err != nil {
			return err
		}

		rBase, err := sta.StatsRegionDatabases.GetRegionDatabase(name)
		if err != nil {
			return err
		}

		logging.WithFields(logrus.Fields{
			"region":           name,
			"connected-realms": len(ids),
			"stats":            len(encodedStats),
			"timestamp":        currentTimestamp,
		}).Info("persisting stats")

		if err := rBase.PersistEncodedStats(currentTimestamp, encodedStats); err != nil {
			return err
		}

		logging.WithFields(logrus.Fields{
			"region":          name,
			"retention-limit": retentionLimit,
		}).Info("pruning stats")
		if err := rBase.PruneStats(retentionLimit); err != nil {
			return err
		}
	}

	logging.WithFields(logrus.Fields{
		"total":          len(regionRealmMap),
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total loaded in region-stats")

	return nil
}

func (sta StatsState) TuplesIntake(tuples blizzardv2.LoadConnectedRealmTuples) error {
	startTime := time.Now()

	// spinning up workers
	getEncodedStatsByTuplesOut := sta.LakeClient.GetEncodedStatsByTuples(tuples)
	persistEncodedStatsIn := make(chan StatsDatabase.PersistRealmStatsInJob)
	persistEncodedStatsOut := sta.StatsTupleDatabases.PersistEncodedRealmStats(persistEncodedStatsIn)

	// loading it in
	go func() {
		for job := range getEncodedStatsByTuplesOut {
			if job.Err() != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to fetch encoded-stats")

				continue
			}

			persistEncodedStatsIn <- StatsDatabase.PersistRealmStatsInJob{
				Tuple:        job.Tuple(),
				EncodedStats: job.EncodedStats(),
			}
		}

		close(persistEncodedStatsIn)
	}()

	// waiting for it to drain out
	totalLoaded := 0
	regionTimestamps := sotah.RegionTimestamps{}
	for job := range persistEncodedStatsOut {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("failed to load encoded stats in")

			return job.Err
		}

		logging.WithFields(logrus.Fields{
			"region":          job.Tuple.RegionName,
			"connected-realm": job.Tuple.ConnectedRealmId,
		}).Info("loaded stats in")

		regionTimestamps = regionTimestamps.SetStatsReceived(
			job.Tuple.RegionConnectedRealmTuple,
			job.Tuple.LastModified,
		)
		totalLoaded += 1
	}

	// optionally updating region state
	if !regionTimestamps.IsZero() {
		if err := sta.ReceiveRegionTimestamps(regionTimestamps); err != nil {
			logging.WithField("error", err.Error()).Error("failed to receive region-timestamps")

			return err
		}
	}

	// pruning stats
	if err := sta.StatsTupleDatabases.PruneRealmStats(
		tuples.RegionConnectedRealmTuples(),
		sotah.UnixTimestamp(BaseDatabase.RetentionLimit().Unix()),
	); err != nil {
		logging.WithField("error", err.Error()).Error("failed to prune stats")

		return err
	}

	logging.WithFields(logrus.Fields{
		"total":          totalLoaded,
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total loaded in tuples-stats")

	return nil
}
