package state

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"

	nats "github.com/nats-io/go-nats"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type RealmModificationDatesResponse struct {
	sotah.ConnectedRealmTimestamps
}

func (r RealmModificationDatesResponse) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(r)
}

func (sta RegionsState) ListenForQueryRealmModificationDates(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.QueryRealmModificationDates), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		req, err := sotah.NewRegionRealmTuple(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		connectedRealmTimestamps, err := sta.RegionComposites.FindRealmTimestamps(req.RegionName, req.RealmSlug)
		if err != nil {
			logging.WithFields(logrus.Fields{
				"region": req.RegionName,
				"realm":  req.RealmSlug,
			}).Error("failed to resolve connected-realm timestamps")

			m.Err = err.Error()
			m.Code = mCodes.UserError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		res := RealmModificationDatesResponse{connectedRealmTimestamps}

		encodedData, err := res.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = string(encodedData)
		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}