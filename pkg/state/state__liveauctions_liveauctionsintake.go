package state

import (
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	LiveAuctionsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/liveauctions" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func NewLiveAuctionsIntakeRequest(data []byte) (LiveAuctionsIntakeRequest, error) {
	out := LiveAuctionsIntakeRequest{}
	if err := json.Unmarshal(data, &out); err != nil {
		return LiveAuctionsIntakeRequest{}, err
	}

	return out, nil
}

type LiveAuctionsIntakeRequest struct {
	Version gameversion.GameVersion               `json:"version"`
	Tuples  blizzardv2.RegionConnectedRealmTuples `json:"tuples"`
}

func (req LiveAuctionsIntakeRequest) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(req)
}

func (sta LiveAuctionsState) ListenForLiveAuctionsIntake(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.LiveAuctionsIntake), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		req, err := NewLiveAuctionsIntakeRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		logging.WithFields(logrus.Fields{
			"version": req.Version,
			"tuples":  len(req.Tuples),
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

func (sta LiveAuctionsState) LiveAuctionsIntake(req LiveAuctionsIntakeRequest) error {
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
		if err := sta.ReceiveRegionTimestamps(req.Version, regionTimestamps); err != nil {
			logging.WithField("error", err.Error()).Error("failed to receive region-timestamps")

			return err
		}
	}

	logging.WithFields(logrus.Fields{
		"total":          totalLoaded,
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total loaded in live-auctions")

	return nil
}
