package state

import (
	"encoding/json"

	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type SessionSecretResponse struct {
	SessionSecret string `json:"session_secret"`
}

func (res SessionSecretResponse) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(res)
}

func (sta BootState) ListenForSessionSecret(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.SessionSecret), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		res := SessionSecretResponse{sta.SessionSecret.String()}

		encodedData, err := res.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
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
