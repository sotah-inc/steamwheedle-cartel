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

func NewRecipeRequest(body []byte) (RecipeRequest, error) {
	out := RecipeRequest{}
	if err := json.Unmarshal(body, &out); err != nil {
		return RecipeRequest{}, err
	}

	return out, nil
}

type RecipeRequest struct {
	RecipeId blizzardv2.RecipeId `json:"recipe_id"`
	Locale   locale.Locale       `json:"locale"`
}

func (req RecipeRequest) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(req)
}

func NewRecipeResponse(base64Encoded string) (RecipeResponse, error) {
	gzipEncoded, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return RecipeResponse{}, err
	}

	gzipDecoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return RecipeResponse{}, err
	}

	out := RecipeResponse{}
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return RecipeResponse{}, err
	}

	return out, nil
}

type RecipeResponse struct {
	Recipe sotah.ShortRecipe `json:"recipe"`
}

func (resp RecipeResponse) EncodeForDelivery() (string, error) {
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

func (sta ProfessionsState) ListenForRecipe(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.Recipe), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		req, err := NewRecipeRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		recipe, err := sta.ProfessionsDatabase.GetRecipe(req.RecipeId)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		resp := RecipeResponse{Recipe: sotah.NewShortRecipe(recipe, req.Locale)}
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
