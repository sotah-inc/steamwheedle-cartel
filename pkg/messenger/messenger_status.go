package messenger

import (
	"encoding/json"
	"errors"

	"git.sotah.info/steamwheedle-cartel/pkg/blizzard"
	"git.sotah.info/steamwheedle-cartel/pkg/messenger/codes"
	"git.sotah.info/steamwheedle-cartel/pkg/sotah"
	"git.sotah.info/steamwheedle-cartel/pkg/state/subjects"
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
	RegionName blizzard.RegionName `json:"region_name"`
}

func (mess Messenger) NewStatus(reg sotah.Region) (sotah.Status, error) {
	lm := StatusRequest{RegionName: reg.Name}
	encodedMessage, err := json.Marshal(lm)
	if err != nil {
		return sotah.Status{}, err
	}

	msg, err := mess.Request(string(subjects.Status), encodedMessage)
	if err != nil {
		return sotah.Status{}, err
	}

	if msg.Code != codes.Ok {
		return sotah.Status{}, errors.New(msg.Err)
	}

	stat, err := blizzard.NewStatus([]byte(msg.Data))
	if err != nil {
		return sotah.Status{}, err
	}

	return sotah.NewStatus(reg, stat), nil
}
