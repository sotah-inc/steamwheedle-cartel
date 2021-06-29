package state

import (
	"fmt"

	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta RegionsState) ListenForStatus(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.Status), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		tuple, err := blizzardv2.NewRegionVersionTuple(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		if !sta.GameVersionList.Includes(tuple.Version) {
			m.Err = "invalid game-version"
			m.Code = codes.UserError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		region, err := sta.RegionsDatabase.GetRegion(tuple.RegionName)
		if err != nil {
			m.Err = fmt.Sprintf("invalid region name: %s", tuple.RegionName)
			m.Code = codes.UserError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		connectedRealms, err := sta.RegionsDatabase.GetConnectedRealms(tuple)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		regionComposite := sotah.RegionComposite{
			ConfigRegion: region,
			ConnectedRealmComposites: map[gameversion.GameVersion]sotah.RealmComposites{
				tuple.Version: connectedRealms,
			},
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
