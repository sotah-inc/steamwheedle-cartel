package state

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/base"
	PricelistHistoryDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/pricelisthistory"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta PricelistHistoryState) ListenForPricelistHistoryIntake(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.PricelistHistoryIntake), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		tuples, err := blizzardv2.NewLoadConnectedRealmTuples(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		logging.WithField("tuples", len(tuples)).Info("received")
		if err := sta.pricelistHistoryIntake(tuples); err != nil {
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

func (sta PricelistHistoryState) pricelistHistoryIntake(tuples blizzardv2.LoadConnectedRealmTuples) error {
	startTime := time.Now()

	// spinning up workers
	getPricelistHistoryByTuplesOut := sta.LakeClient.GetEncodedPricelistHistoryByTuples(tuples)
	loadEncodedDataIn := make(chan PricelistHistoryDatabase.LoadEncodedDataInJob)
	loadEncodedDataOut := sta.PricelistHistoryDatabases.LoadEncodedData(loadEncodedDataIn)

	// loading it in
	go func() {
		for job := range getPricelistHistoryByTuplesOut {
			if job.Err() != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to fetch pricelist-history")

				continue
			}

			loadEncodedDataIn <- PricelistHistoryDatabase.LoadEncodedDataInJob{
				Tuple:       job.Tuple(),
				EncodedData: job.EncodedPricelistHistory(),
			}
		}

		close(loadEncodedDataIn)
	}()

	// waiting for it to drain out
	totalLoaded := 0
	regionTimestamps := sotah.RegionTimestamps{}
	for job := range loadEncodedDataOut {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("failed to load encoded pricelist-history in")

			return job.Err
		}

		logging.WithFields(logrus.Fields{
			"region":          job.Tuple.RegionName,
			"connected-realm": job.Tuple.ConnectedRealmId,
		}).Info("loaded pricelist-history in")

		regionTimestamps = regionTimestamps.SetPricelistHistoryReceived(
			job.Tuple.RegionConnectedRealmTuple,
			job.ReceivedAt,
		)

		totalLoaded += 1
	}

	// optionally updating region state
	if !regionTimestamps.IsZero() {
		sta.ReceiveRegionTimestamps(regionTimestamps)
	}

	// pruning databases where applicable
	if err := sta.PricelistHistoryDatabases.PruneDatabases(
		sotah.UnixTimestamp(BaseDatabase.RetentionLimit().Unix()),
	); err != nil {
		logging.WithField("error", err.Error()).Error("failed to prune pricelist-history databases")

		return err
	}

	logging.WithFields(logrus.Fields{
		"total":          totalLoaded,
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total loaded in pricelist-history")

	return nil
}
