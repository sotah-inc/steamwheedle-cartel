package state

import (
	"encoding/json"

	nats "github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type QueryRealmModificationDatesResponse struct {
	sotah.StatusTimestamps
}

func (r QueryRealmModificationDatesResponse) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(r)
}

func (sta RegionsState) ListenForQueryRealmModificationDates(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(
		string(subjects.QueryRealmModificationDates),
		stop,
		func(natsMsg nats.Msg) {
			m := messenger.NewMessage()

			tuple, err := blizzardv2.NewRegionVersionConnectedRealmTuple(natsMsg.Data)
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			if !sta.GameVersionList.Includes(tuple.Version) {
				m.Err = "invalid game-version"
				m.Code = mCodes.UserError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			connectedRealmTimestamps, err := sta.RegionsDatabase.GetStatusTimestamps(tuple)
			if err != nil {
				logging.WithFields(logrus.Fields{
					"region":          tuple.RegionName,
					"game-version":    tuple.Version,
					"connected-realm": tuple.ConnectedRealmId,
				}).Error("failed to resolve connected-realm timestamps")

				m.Err = err.Error()
				m.Code = mCodes.NotFound
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			res := QueryRealmModificationDatesResponse{connectedRealmTimestamps}

			encodedData, err := res.EncodeForDelivery()
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			m.Data = string(encodedData)
			sta.Messenger.ReplyTo(natsMsg, m)
		},
	)
	if err != nil {
		return err
	}

	return nil
}
