package state

import (
	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type ValidateRegionRealmResponse struct {
	ValidateRegionConnectedRealmResponse
}

func (sta RegionsState) ListenForValidateRegionRealm(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.ValidateRegionRealm), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		req, err := blizzardv2.NewRegionRealmTuple(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		res := ValidateRegionRealmResponse{
			ValidateRegionConnectedRealmResponse{
				IsValid: sta.RegionComposites.RegionRealmExists(req.RegionName, req.RealmSlug),
			},
		}
		encoded, err := res.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = encoded
		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
