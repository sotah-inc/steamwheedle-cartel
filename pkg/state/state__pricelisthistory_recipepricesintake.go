package state

import (
	"errors"
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

func (sta PricelistHistoryState) ListenForRecipePricesIntake(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.RecipePricesIntake), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		tuples, err := blizzardv2.NewLoadConnectedRealmTuples(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		logging.WithField("tuples", len(tuples)).Info("received")
		if err := sta.recipePricesIntake(tuples); err != nil {
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

func (sta PricelistHistoryState) recipePricesIntake(
	tuples blizzardv2.LoadConnectedRealmTuples,
) error {
	startTime := time.Now()

	respMsg, err := sta.Messenger.Request(messenger.RequestOptions{
		Subject: string(subjects.MiniRecipes),
		Data:    nil,
		Timeout: 1 * time.Minute,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to gather mini-recipes")

		return err
	}
	if respMsg.Code != codes.Ok {
		logging.WithField("code", respMsg.Code).Error("failed to gather mini-recipes")

		return errors.New("mini-recipes code was not ok")
	}

	mRecipesResponse, err := NewMiniRecipesResponse(respMsg.Data)
	if err != nil {
		return err
	}

	// spinning up workers
	getEncodedRecipePricesOut := sta.LakeClient.GetEncodedRecipePricesByTuples(
		mRecipesResponse.Recipes,
		tuples,
	)
	loadEncodedRecipePricesIn := make(chan PricelistHistoryDatabase.LoadEncodedRecipePricesInJob)
	loadEncodedRecipePricesOut := sta.PricelistHistoryDatabases.LoadEncodedRecipePrices(
		loadEncodedRecipePricesIn,
	)

	// loading it in
	go func() {
		for job := range getEncodedRecipePricesOut {
			if job.Err() != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to fetch encoded recipe-prices")

				continue
			}

			loadEncodedRecipePricesIn <- PricelistHistoryDatabase.LoadEncodedRecipePricesInJob{
				Tuple:       job.Tuple(),
				EncodedData: job.EncodedRecipePrices(),
			}
		}

		close(loadEncodedRecipePricesIn)
	}()

	// waiting for it to drain out
	totalLoaded := 0
	regionTimestamps := sotah.RegionVersionTimestamps{}
	for job := range loadEncodedRecipePricesOut {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("failed to load encoded recipe-prices in")

			return job.Err
		}

		logging.WithFields(logrus.Fields{
			"region":          job.Tuple.RegionName,
			"connected-realm": job.Tuple.ConnectedRealmId,
		}).Info("loaded encoded recipe-prices in")

		regionTimestamps = regionTimestamps.SetTimestamp(
			job.Tuple.RegionVersionConnectedRealmTuple,
			statuskinds.RecipePricesReceived,
			job.ReceivedAt,
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

	logging.WithFields(logrus.Fields{
		"total":          totalLoaded,
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total loaded in encoded recipe-prices")

	return nil
}
