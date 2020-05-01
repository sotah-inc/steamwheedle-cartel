package state

import (
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
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
		if err := sta.PricelistHistoryIntake(tuples); err != nil {
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

func (sta PricelistHistoryState) PricelistHistoryIntake(tuples blizzardv2.LoadConnectedRealmTuples) error {
	// spinning up workers
	getPricelistHistoryByTuplesOut := sta.LakeClient.GetEncodedPricelistHistoryByTuples(tuples)
	loadEncodedDataIn := make(chan database.PricelistHistoryLoadEncodedDataInJob)
	loadEncodedDataOut := sta.PricelistHistoryDatabases.LoadEncodedData(loadEncodedDataIn)

	// loading it in
	go func() {
		for job := range getPricelistHistoryByTuplesOut {
			if job.Err() != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to fetch pricelist-history")

				continue
			}

			loadEncodedDataIn <- database.PricelistHistoryLoadEncodedDataInJob{
				Tuple:       job.Tuple(),
				EncodedData: job.EncodedPricelistHistory(),
			}
		}

		close(loadEncodedDataIn)
	}()

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
	}

	// optionally updating region state
	if !regionTimestamps.IsZero() {
		sta.ReceiveRegionTimestamps(regionTimestamps)
	}

	// pruning databases where applicable
	if err := sta.PricelistHistoryDatabases.PruneDatabases(
		sotah.UnixTimestamp(database.RetentionLimit().Unix()),
	); err != nil {
		logging.WithField("error", err.Error()).Error("failed to prune pricelist-history databases")

		return err
	}

	return nil
}
