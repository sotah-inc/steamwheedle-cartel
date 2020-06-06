package state

import (
	"fmt"

	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta RegionsState) ListenForConnectedRealms(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.ConnectedRealms), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		sRequest, err := blizzardv2.NewRegionTuple(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		region, err := sta.RegionComposites.FindByRegionName(sRequest.RegionName)
		if err != nil {
			m.Err = fmt.Sprintf("invalid region name: %s", sRequest.RegionName)
			m.Code = codes.NotFound
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		encoded, err := region.ConnectedRealmComposites.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = string(encoded)
		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
