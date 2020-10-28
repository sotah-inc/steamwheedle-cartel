package state

import (
	"time"

	LiveAuctionsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/liveauctions"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta LiveAuctionsState) ListenForLiveAuctionsIntake(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.LiveAuctionsIntake), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		tuples, err := blizzardv2.NewLoadConnectedRealmTuples(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		logging.WithField("tuples", len(tuples)).Info("received")
		if err := sta.LiveAuctionsIntake(tuples); err != nil {
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

func (sta LiveAuctionsState) LiveAuctionsIntake(tuples blizzardv2.LoadConnectedRealmTuples) error {
	startTime := time.Now()

	// spinning up workers
	getAuctionsByTuplesOut := sta.LakeClient.GetEncodedAuctionsByTuples(tuples.RegionConnectedRealmTuples())
	loadEncodedDataIn := make(chan LiveAuctionsDatabase.LiveAuctionsLoadEncodedDataInJob)
	loadEncodedDataOut := sta.LiveAuctionsDatabases.LoadEncodedData(loadEncodedDataIn)

	// loading it in
	go func() {
		for job := range getAuctionsByTuplesOut {
			if job.Err() != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to fetch auctions")

				continue
			}

			loadEncodedDataIn <- LiveAuctionsDatabase.LiveAuctionsLoadEncodedDataInJob{
				Tuple:       job.Tuple(),
				EncodedData: job.EncodedAuctions(),
			}
		}

		close(loadEncodedDataIn)
	}()

	// waiting for it to drain out
	totalLoaded := 0
	regionTimestamps := sotah.RegionTimestamps{}
	for job := range loadEncodedDataOut {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("failed to load encoded auctions in")

			return job.Err
		}

		logging.WithFields(logrus.Fields{
			"region":          job.Tuple.RegionName,
			"connected-realm": job.Tuple.ConnectedRealmId,
		}).Info("loaded auctions in")

		regionTimestamps = regionTimestamps.SetLiveAuctionsReceived(job.Tuple, job.ReceivedAt)
		totalLoaded += 1
	}

	// optionally updating region state
	if !regionTimestamps.IsZero() {
		sta.ReceiveRegionTimestamps(regionTimestamps)
	}

	// persisting related stats
	if err := sta.LiveAuctionsDatabases.PersistStats(tuples.RegionConnectedRealmTuples()); err != nil {
		logging.WithField("error", err.Error()).Error("failed to persist live-auctions stats")

		return err
	}

	// pruning stats
	if err := sta.LiveAuctionsDatabases.PruneStats(
		tuples.RegionConnectedRealmTuples(),
		sotah.UnixTimestamp(BaseDatabase.RetentionLimit().Unix()),
	); err != nil {
		logging.WithField("error", err.Error()).Error("failed to prune live-auctions stats")

		return err
	}

	logging.WithFields(logrus.Fields{
		"total":          totalLoaded,
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total loaded in live-auctions")

	return nil
}
