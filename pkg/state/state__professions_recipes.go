package state

import (
	"encoding/base64"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/professions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewRecipesRequest(body []byte) (RecipesRequest, error) {
	out := RecipesRequest{}
	if err := json.Unmarshal(body, &out); err != nil {
		return RecipesRequest{}, err
	}

	return out, nil
}

type RecipesRequest struct {
	RecipeIds []blizzardv2.RecipeId `json:"recipe_ids"`
	Locale    locale.Locale         `json:"locale"`
}

func (req RecipesRequest) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(req)
}

func NewRecipesResponse(base64Encoded string) (RecipesResponse, error) {
	gzipEncoded, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return RecipesResponse{}, err
	}

	gzipDecoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return RecipesResponse{}, err
	}

	out := RecipesResponse{}
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return RecipesResponse{}, err
	}

	return out, nil
}

type RecipesResponse struct {
	Recipes []sotah.ShortRecipe `json:"recipes"`
}

func (resp RecipesResponse) EncodeForDelivery() (string, error) {
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

func (sta ProfessionsState) ListenForRecipes(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.Recipes), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		req, err := NewRecipesRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		recipesOut := sta.ProfessionsDatabase.GetRecipes(req.RecipeIds)
		var recipes []sotah.Recipe
		for recipesOutJob := range recipesOut {
			if recipesOutJob.Err != nil {
				if getRecipeError, ok := recipesOutJob.Err.(*professions.GetRecipeError); ok {
					if !getRecipeError.Exists {
						continue
					}
				}

				m.Err = recipesOutJob.Err.Error()
				m.Code = codes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			recipes = append(recipes, recipesOutJob.Recipe)
		}

		respRecipes := make([]sotah.ShortRecipe, len(recipes))
		for i, recipe := range recipes {
			respRecipes[i] = sotah.NewShortRecipe(recipe, req.Locale)
		}

		resp := RecipesResponse{Recipes: respRecipes}
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
