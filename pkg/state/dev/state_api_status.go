package dev

import (
	"encoding/json"
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	nats "github.com/nats-io/go-nats"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func NewStatusRequest(payload []byte) (StatusRequest, error) {
	sr := &StatusRequest{}
	err := json.Unmarshal(payload, &sr)
	if err != nil {
		return StatusRequest{}, err
	}

	return *sr, nil
}

type StatusRequest struct {
	RegionName blizzardv2.RegionName `json:"region_name"`
}

func (sta *APIState) ListenForStatus(stop state.ListenStopChan) error {
	err := sta.messenger.Subscribe(string(subjects.Status), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		sRequest, err := NewStatusRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.messenger.ReplyTo(natsMsg, m)

			return
		}

		foundConnectedRealms, ok := sta.regionConnectedRealms[sRequest.RegionName]
		if !ok {
			m.Err = fmt.Sprintf("invalid region name: %s", sRequest.RegionName)
			m.Code = codes.UserError
			sta.messenger.ReplyTo(natsMsg, m)

			return
		}

		encodedStatus, err := func() ([]byte, error) {
			jsonEncoded, err := json.Marshal(foundConnectedRealms)
			if err != nil {
				return []byte{}, err
			}

			return util.GzipEncode(jsonEncoded)
		}()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = string(encodedStatus)
		sta.messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
