package dev

import (
	"encoding/json"

	"git.sotah.info/steamwheedle-cartel/pkg/messenger"
	"git.sotah.info/steamwheedle-cartel/pkg/messenger/codes"
	"git.sotah.info/steamwheedle-cartel/pkg/state"
	"git.sotah.info/steamwheedle-cartel/pkg/state/subjects"
	nats "github.com/nats-io/go-nats"
)

func (sta APIState) ListenForBoot(stop state.ListenStopChan) error {
	err := sta.IO.Messenger.Subscribe(string(subjects.Boot), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		encodedResponse, err := json.Marshal(state.BootResponse{
			Regions:     sta.Regions,
			ItemClasses: sta.ItemClasses,
			Expansions:  sta.Expansions,
			Professions: sta.Professions,
		})
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = string(encodedResponse)
		sta.IO.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
