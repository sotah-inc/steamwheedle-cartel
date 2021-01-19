package state

import (
	"encoding/json"

	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type ConnectedRealmModificationDatesResponse map[blizzardv2.ConnectedRealmId]sotah.ConnectedRealmTimestamps // nolint:lll

func (r ConnectedRealmModificationDatesResponse) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(r)
}

func (sta RegionsState) ListenForConnectedRealmModificationDates(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(
		string(subjects.ConnectedRealmModificationDates),
		stop,
		func(natsMsg nats.Msg) {
			m := messenger.NewMessage()

			req, err := blizzardv2.NewRegionTuple(natsMsg.Data)
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			foundTimestamps, err := sta.RegionTimestamps().FindByRegionName(req.RegionName)
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.UserError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			res := ConnectedRealmModificationDatesResponse(foundTimestamps)
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
