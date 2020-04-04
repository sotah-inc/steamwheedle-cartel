package state

import (
	"encoding/json"

	nats "github.com/nats-io/go-nats"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func NewValidateRegionRealmRequest(data []byte) (ValidateRegionRealmRequest, error) {
	var out ValidateRegionRealmRequest
	if err := json.Unmarshal(data, &out); err != nil {
		return ValidateRegionRealmRequest{}, err
	}

	return out, nil
}

type ValidateRegionRealmRequest struct {
	RegionName blizzardv2.RegionName `json:"region_name"`
	RealmSlug  blizzardv2.RealmSlug  `json:"realm_slug"`
}

type ValidateRegionRealmResponse struct {
	IsValid bool `json:"is_valid"`
}

func (res ValidateRegionRealmResponse) EncodeForDelivery() (string, error) {
	encodedResult, err := json.Marshal(res)
	if err != nil {
		return "", err
	}

	return string(encodedResult), nil
}

func (sta RegionsState) ListenForValidateRegionRealm(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.ValidateRegionRealm), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		req, err := NewValidateRegionRealmRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		res := ValidateRegionRealmResponse{
			IsValid: sta.RegionComposites.RegionRealmExists(req.RegionName, req.RealmSlug),
		}
		encoded, err := res.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = string(encoded)
		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
