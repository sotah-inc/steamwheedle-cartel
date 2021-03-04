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

func NewProfessionsFromIdsRequest(body []byte) (ProfessionsFromIdsRequest, error) {
	out := ProfessionsFromIdsRequest{}
	if err := json.Unmarshal(body, &out); err != nil {
		return ProfessionsFromIdsRequest{}, err
	}

	return out, nil
}

type ProfessionsFromIdsRequest struct {
	ProfessionIds []blizzardv2.ProfessionId `json:"profession_ids"`
	Locale        locale.Locale             `json:"locale"`
}

func (req ProfessionsFromIdsRequest) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(req)
}

func NewProfessionsFromIdsResponse(base64Encoded string) (ProfessionsFromIdsResponse, error) {
	gzipEncoded, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return ProfessionsFromIdsResponse{}, err
	}

	gzipDecoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return ProfessionsFromIdsResponse{}, err
	}

	out := ProfessionsFromIdsResponse{}
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return ProfessionsFromIdsResponse{}, err
	}

	return out, nil
}

type ProfessionsFromIdsResponse struct {
	Professions sotah.ShortProfessions `json:"professions"`
}

func (resp ProfessionsFromIdsResponse) EncodeForDelivery() (string, error) {
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

func (sta ProfessionsState) ListenForProfessionsFromIds(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.ProfessionsFromIds), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		req, err := NewProfessionsFromIdsRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		professions, err := sta.ProfessionsDatabase.GetProfessionsFromIds(req.ProfessionIds)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		respProfessions := make(sotah.ShortProfessions, len(professions))
		for i, profession := range professions {
			respProfessions[i] = sotah.NewShortProfession(profession, req.Locale)
		}

		resp := ProfessionsFromIdsResponse{Professions: respProfessions}
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
