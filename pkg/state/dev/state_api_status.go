package dev

import (
	"encoding/json"

	"github.com/nats-io/go-nats"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/messenger"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/messenger/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
)

func (sta APIState) ListenForStatus(stop state.ListenStopChan) error {
	err := sta.IO.Messenger.Subscribe(string(subjects.Status), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		sr, err := messenger.NewStatusRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		reg, err := sta.Regions.GetRegion(sr.RegionName)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.NotFound
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		regionStatus, ok := sta.Statuses[reg.Name]
		if !ok {
			m.Err = "Region found but not in Statuses"
			m.Code = codes.NotFound
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		encodedStatus, err := json.Marshal(regionStatus)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = string(encodedStatus)
		sta.IO.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
