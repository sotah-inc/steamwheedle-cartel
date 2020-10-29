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

type BootResponse struct {
	Regions     sotah.RegionList               `json:"regions"`
	ItemClasses []blizzardv2.ItemClassResponse `json:"item_classes"`
	Expansions  []sotah.Expansion              `json:"expansions"`
	Professions []sotah.ConfigProfession       `json:"professions"`
}

func (res BootResponse) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(res)
}

func (sta BootState) ListenForBoot(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.Boot), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		res := BootResponse{
			Regions:     sta.Regions,
			ItemClasses: sta.ItemClasses,
			Expansions:  sta.Expansions,
			Professions: sta.Professions,
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
