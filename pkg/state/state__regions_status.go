package state

import (
	"fmt"

	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta RegionsState) ListenForStatus(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.Status), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		sRequest, err := blizzardv2.NewVersionRegionTuple(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		if !sta.GameVersionList.Includes(sRequest.Version) {
			m.Err = "invalid game-version"
			m.Code = codes.UserError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		region, err := sta.RegionsDatabase.GetRegion(sRequest.RegionName)
		if err != nil {
			m.Err = fmt.Sprintf("invalid region name: %s", sRequest.RegionName)
			m.Code = codes.UserError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		realmComposites, err := sta.RegionsDatabase.GetConnectedRealms(
			sRequest.Version,
			sRequest.RegionName,
		)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		regionComposite := sotah.RegionComposite{
			ConfigRegion:             region,
			ConnectedRealmComposites: realmComposites,
		}

		encodedStatus, err := regionComposite.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = encodedStatus
		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
