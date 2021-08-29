package state

import (
	"encoding/base64"
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"

	nats "github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func NewItemsFindMatchingRecipesRequest(data string) (ItemsFindMatchingRecipesRequest, error) {
	gzipEncoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return ItemsFindMatchingRecipesRequest{}, err
	}

	jsonEncoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return ItemsFindMatchingRecipesRequest{}, err
	}

	out := ItemsFindMatchingRecipesRequest{}

	if err := json.Unmarshal(jsonEncoded, &out); err != nil {
		return ItemsFindMatchingRecipesRequest{}, err
	}

	return out, nil
}

type ItemsFindMatchingRecipesRequest struct {
	Version    gameversion.GameVersion     `json:"game_version"`
	RecipesMap blizzardv2.RecipeSubjectMap `json:"recipes_map"`
}

func (req ItemsFindMatchingRecipesRequest) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}

func (sta ItemsState) ListenForItemsFindMatchingRecipes(stop ListenStopChan) error {
	return sta.Messenger.Subscribe(
		string(subjects.ItemsFindMatchingRecipes),
		stop,
		func(natsMsg nats.Msg) {
			m := messenger.NewMessage()

			logging.Info("handling request for items find-matching-recipes")

			req, err := NewItemsFindMatchingRecipesRequest(string(natsMsg.Data))
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			logging.WithFields(logrus.Fields{
				"game-version":    req.Version,
				"recipe-subjects": len(req.RecipesMap),
			}).Info("decoded recipe-subjects from request")

			irMap, err := sta.ItemsDatabase.FindMatchingItemsFromRecipes(req.Version, req.RecipesMap)
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			logging.WithFields(logrus.Fields{
				"game-version":    req.Version,
				"recipe-subjects": len(req.RecipesMap),
				"item-recipes":    len(irMap),
			}).Info("found matching item-recipes with recipe-subjects")

			encodedItemRecipes, err := irMap.EncodeForDelivery()
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			// dumping it out
			m.Data = encodedItemRecipes
			sta.Messenger.ReplyTo(natsMsg, m)
		},
	)
}
