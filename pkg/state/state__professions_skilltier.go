package state

import (
	"encoding/base64"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewSkillTierRequest(body []byte) (SkillTierRequest, error) {
	out := SkillTierRequest{}
	if err := json.Unmarshal(body, &out); err != nil {
		return SkillTierRequest{}, err
	}

	return out, nil
}

type SkillTierRequest struct {
	ProfessionId blizzardv2.ProfessionId `json:"profession_id"`
	SkillTierId  blizzardv2.SkillTierId  `json:"skilltier_id"`
	Locale       locale.Locale           `json:"locale"`
}

func (req SkillTierRequest) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(req)
}

func NewSkillTierResponse(base64Encoded string) (SkillTierResponse, error) {
	gzipEncoded, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return SkillTierResponse{}, err
	}

	gzipDecoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return SkillTierResponse{}, err
	}

	out := SkillTierResponse{}
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return SkillTierResponse{}, err
	}

	return out, nil
}

type SkillTierResponse struct {
	SkillTier sotah.ShortSkillTier `json:"skilltier"`
}

func (resp SkillTierResponse) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}

func (sta ProfessionsState) ListenForSkillTier(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.SkillTier), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		req, err := NewSkillTierRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		skillTier, err := sta.ProfessionsDatabase.GetSkillTier(req.ProfessionId, req.SkillTierId)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		resp := SkillTierResponse{SkillTier: sotah.NewShortSkillTier(skillTier, req.Locale)}
		data, err := resp.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = data

		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
