package state

import (
	"encoding/json"

	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type ResolveConnectedRealmResponse struct {
	ConnectedRealm sotah.RealmComposite `json:"connected_realm"`
}

func (res ResolveConnectedRealmResponse) EncodeForDelivery() (string, error) {
	encodedResult, err := json.Marshal(res)
	if err != nil {
		return "", err
	}

	return string(encodedResult), nil
}

func (sta RegionsState) ListenForResolveConnectedRealm(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(
		string(subjects.ResolveConnectedRealm),
		stop,
		func(natsMsg nats.Msg) {
			m := messenger.NewMessage()

			tuple, err := blizzardv2.NewRegionVersionRealmTuple(natsMsg.Data)
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

			connectedRealm, err := sta.RegionsDatabase.GetConnectedRealmBySlugTuple(tuple)
			if err != nil {
				m.Err = err.Error()
				m.Code = codes.NotFound
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			res := ResolveConnectedRealmResponse{
				ConnectedRealm: connectedRealm,
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
		},
	)
	if err != nil {
		return err
	}

	return nil
}
