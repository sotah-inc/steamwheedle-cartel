package state

import (
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta LiveAuctionsState) ListenForLiveAuctionsIntake(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.LiveAuctionsIntake), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		tuples, err := blizzardv2.NewRegionConnectedRealmTuples(natsMsg.Data)
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

func (sta LiveAuctionsState) LiveAuctionsIntake(tuples blizzardv2.RegionConnectedRealmTuples) error {
	// spinning up workers
	getAuctionsByTuplesOut := sta.DiskStore.GetAuctionsByTuples(tuples)
	loadEncodedDataIn := make(chan database.LiveAuctionsLoadEncodedDataInJob)
	loadEncodedDataOut := sta.LiveAuctionsDatabases.LoadEncodedData(loadEncodedDataIn)

	// loading it in
	go func() {
		for job := range getAuctionsByTuplesOut {
			if job.Err != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to fetch auctions")

				continue
			}

			encodedData, err := job.Auctions.EncodeForStorage()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"err":             err.Error(),
					"region":          job.Tuple.RegionName,
					"connected-realm": job.Tuple.ConnectedRealmId,
				}).Error("failed to encode auctions for storage")

				continue
			}

			loadEncodedDataIn <- database.LiveAuctionsLoadEncodedDataInJob{
				Tuple:       job.Tuple,
				EncodedData: encodedData,
			}
		}

		close(loadEncodedDataIn)
	}()

	for job := range loadEncodedDataOut {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("failed to load encoded auctions in")

			return job.Err
		}

		logging.WithFields(logrus.Fields{
			"region":          job.Tuple.RegionName,
			"connected-realm": job.Tuple.ConnectedRealmId,
		}).Info("loaded auctions in")
	}

	return nil
}