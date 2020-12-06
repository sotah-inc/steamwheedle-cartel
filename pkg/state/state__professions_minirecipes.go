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

func NewMiniRecipesResponse(base64Encoded string) (MiniRecipesResponse, error) {
	base64Decoded, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return MiniRecipesResponse{}, err
	}

	gzipDecoded, err := util.GzipDecode(base64Decoded)
	if err != nil {
		return MiniRecipesResponse{}, err
	}

	out := MiniRecipesResponse{}
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return MiniRecipesResponse{}, err
	}

	return out, nil
}

type MiniRecipesResponse struct {
	Recipes sotah.MiniRecipes `json:"recipes"`
}

func (resp MiniRecipesResponse) EncodeForDelivery() (string, error) {
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

func (sta ProfessionsState) ListenForMiniRecipes(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.MiniRecipes), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		mRecipes, err := sta.ProfessionsDatabase.GetMiniRecipes()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		resp := MiniRecipesResponse{Recipes: mRecipes}
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
