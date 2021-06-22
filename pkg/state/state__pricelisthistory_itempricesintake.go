package state

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	PricelistHistoryDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/pricelisthistory" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/statuskinds"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta PricelistHistoryState) ListenForItemPricesIntake(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.ItemPricesIntake), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		tuples, err := blizzardv2.NewLoadConnectedRealmTuples(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		logging.WithField("tuples", len(tuples)).Info("received")
		if err := sta.itemPricesIntake(tuples); err != nil {
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

func (sta PricelistHistoryState) itemPricesIntake(
	tuples blizzardv2.LoadConnectedRealmTuples,
) error {
	startTime := time.Now()

	// spinning up workers
	getEncodedItemPricesOut := sta.LakeClient.GetEncodedItemPricesByTuples(tuples)
	loadEncodedDataIn := make(chan PricelistHistoryDatabase.LoadEncodedItemPricesInJob)
	loadEncodedDataOut := sta.PricelistHistoryDatabases.LoadEncodedItemPrices(loadEncodedDataIn)

	// loading it in
	go func() {
		for job := range getEncodedItemPricesOut {
			if job.Err() != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to fetch encoded item-prices")

				continue
			}

			loadEncodedDataIn <- PricelistHistoryDatabase.LoadEncodedItemPricesInJob{
				Tuple:       job.Tuple(),
				EncodedData: job.EncodedPricelistHistory(),
			}
		}

		close(loadEncodedDataIn)
	}()

	// waiting for it to drain out
	totalLoaded := 0
	regionTimestamps := sotah.RegionVersionTimestamps{}
	for job := range loadEncodedDataOut {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("failed to load encoded item-prices in")

			return job.Err
		}

		logging.WithFields(logrus.Fields{
			"region":          job.Tuple.RegionName,
			"connected-realm": job.Tuple.ConnectedRealmId,
		}).Info("loaded encoded item-prices in")

		regionTimestamps = regionTimestamps.SetTimestamp(
			job.Tuple.RegionVersionConnectedRealmTuple,
			statuskinds.ItemPricesReceived,
			job.ReceivedAt,
		)

		totalLoaded += 1
	}

	// optionally updating region state
	if !regionTimestamps.IsZero() {
		if err := sta.ReceiveRegionTimestamps(regionTimestamps); err != nil {
			logging.WithField("error", err.Error()).Error("failed to receive region-timestamps")

			logging.WithField("error", err.Error()).Error("failed to receive region-timestamps")

			return err
		}
	}

	logging.WithFields(logrus.Fields{
		"total":          totalLoaded,
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total loaded in encoded item-prices")

	return nil
}
