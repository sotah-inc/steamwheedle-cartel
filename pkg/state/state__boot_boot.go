package state

import (
	"encoding/json"

	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type BootResponse struct {
	Regions        sotah.RegionList     `json:"regions"`
	FirebaseConfig sotah.FirebaseConfig `json:"firebase_config"`
	VersionMeta    []sotah.VersionMeta  `json:"version_meta"`
}

func (res BootResponse) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(res)
}

func (sta BootState) ListenForBoot(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.Boot), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		res := BootResponse{
			Regions:        sta.Regions,
			FirebaseConfig: sta.FirebaseConfig,
			VersionMeta:    sta.VersionMeta,
		}

		encodedResponse, err := res.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = string(encodedResponse)
		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
