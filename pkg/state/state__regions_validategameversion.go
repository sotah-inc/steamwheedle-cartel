package state

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"

	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type ValidateGameVersionResponse struct {
	ValidateRegionConnectedRealmResponse
}

func (res ValidateGameVersionResponse) EncodeForDelivery() (string, error) {
	encodedResult, err := json.Marshal(res)
	if err != nil {
		return "", err
	}

	return string(encodedResult), nil
}

func (sta RegionsState) ListenForValidateGameVersion(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(
		string(subjects.ValidateGameVersion),
		stop,
		func(natsMsg nats.Msg) {
			m := messenger.NewMessage()

			res := ValidateGameVersionResponse{
				ValidateRegionConnectedRealmResponse{
					IsValid: sta.GameVersionList.Includes(gameversion.GameVersion(natsMsg.Data)),
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
		},
	)
	if err != nil {
		return err
	}

	return nil
}
