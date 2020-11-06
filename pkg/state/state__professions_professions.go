package state

import (
	"encoding/base64"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewProfessionsResponse(base64Encoded string) (ProfessionsResponse, error) {
	gzipEncoded, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return ProfessionsResponse{}, err
	}

	gzipDecoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return ProfessionsResponse{}, err
	}

	out := ProfessionsResponse{}
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return ProfessionsResponse{}, err
	}

	return out, nil
}

type ProfessionsResponse struct {
	Professions []sotah.ShortProfession `json:"professions"`
}

func (resp ProfessionsResponse) EncodeForDelivery() (string, error) {
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

func (sta ProfessionsState) ListenForProfessions(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.Professions), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		professions, err := sta.ProfessionsDatabase.GetProfessions()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		resp := ProfessionsResponse{Professions: sotah.NewShortProfessions(professions)}
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
