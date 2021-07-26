package state

import (
	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type ValidateRegionResponse struct {
	ValidateRegionConnectedRealmResponse
}

func (sta RegionsState) ListenForValidateRegion(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.ValidateRegion), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		tuple, err := blizzardv2.NewRegionTuple(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		exists, err := sta.RegionsDatabase.RegionExists(tuple.RegionName)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		res := ValidateRegionResponse{ValidateRegionConnectedRealmResponse{IsValid: exists}}
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
