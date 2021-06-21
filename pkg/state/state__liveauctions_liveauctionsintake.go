package state

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/statuskinds"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	LiveAuctionsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/liveauctions" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta LiveAuctionsState) ListenForLiveAuctionsIntake(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.LiveAuctionsIntake), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		req, err := NewIntakeRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		logging.WithFields(logrus.Fields{
			"tuples": len(req.Tuples),
		}).Info("received")
		if err := sta.LiveAuctionsIntake(req); err != nil {
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

func (sta LiveAuctionsState) LiveAuctionsIntake(req IntakeRequest) error {
	startTime := time.Now()

	// spinning up workers
	getAuctionsByTuplesOut := sta.LakeClient.GetEncodedAuctionsByTuples(req.Tuples)
	loadEncodedDataIn := make(chan LiveAuctionsDatabase.LoadEncodedDataInJob)
	loadEncodedDataOut := sta.LiveAuctionsDatabases.LoadEncodedData(loadEncodedDataIn)

	// loading it in
	go func() {
		for job := range getAuctionsByTuplesOut {
			if job.Err() != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to fetch auctions")

				continue
			}

			loadEncodedDataIn <- LiveAuctionsDatabase.LoadEncodedDataInJob{
				Tuple:       job.Tuple(),
				EncodedData: job.EncodedAuctions(),
			}
		}

		close(loadEncodedDataIn)
	}()

	// waiting for it to drain out
	totalLoaded := 0
	regionVersionTimestamps := sotah.RegionVersionTimestamps{}
	for job := range loadEncodedDataOut {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("failed to load encoded auctions in")

			return job.Err
		}

		logging.WithFields(logrus.Fields{
			"region":          job.Tuple.RegionName,
			"connected-realm": job.Tuple.ConnectedRealmId,
		}).Info("loaded auctions in")

		regionVersionTimestamps = regionVersionTimestamps.SetTimestamp(
			job.Tuple,
			statuskinds.LiveAuctionsReceived,
			job.ReceivedAt,
		)
		totalLoaded += 1
	}

	// optionally updating region state
	if !regionVersionTimestamps.IsZero() {
		if err := sta.ReceiveRegionTimestamps(regionVersionTimestamps); err != nil {
			logging.WithField("error", err.Error()).Error("failed to receive region-version-timestamps")

			return err
		}
	}

	logging.WithFields(logrus.Fields{
		"total":          totalLoaded,
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total loaded in live-auctions")

	return nil
}
