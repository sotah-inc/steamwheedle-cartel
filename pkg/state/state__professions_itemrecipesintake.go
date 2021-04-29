package state

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/professions/itemrecipekind" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewItemRecipesIntakeRequest(base64Encoded string) (ItemRecipesIntakeRequest, error) {
	gzipEncoded, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return ItemRecipesIntakeRequest{}, err
	}

	jsonEncoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return ItemRecipesIntakeRequest{}, err
	}

	out := ItemRecipesIntakeRequest{}
	if err := json.Unmarshal(jsonEncoded, &out); err != nil {
		return ItemRecipesIntakeRequest{}, err
	}

	return out, nil
}

type ItemRecipesIntakeRequest struct {
	Kind           itemrecipekind.ItemRecipeKind `json:"kind"`
	ItemRecipesMap blizzardv2.ItemRecipesMap     `json:"item_recipes"`
}

func (req ItemRecipesIntakeRequest) EncodeForDelivery() (string, error) {
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

func (sta ProfessionsState) ListenForItemRecipesIntake(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.ItemRecipesIntake), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		req, err := NewItemRecipesIntakeRequest(string(natsMsg.Data))
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		if err := sta.ItemRecipesIntake(req); err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}

func (sta ProfessionsState) ItemRecipesIntake(req ItemRecipesIntakeRequest) error {
	startTime := time.Now()

	logging.WithFields(logrus.Fields{
		"item-recipes": len(req.ItemRecipesMap),
	}).Info("handling request for professions item-recipes intake")

	// resolving existing item-recipes and merging results in
	currentIrMap, err := sta.ProfessionsDatabase.GetItemRecipesMap(
		req.Kind,
		req.ItemRecipesMap.ItemIds(),
	)
	if err != nil {
		logging.WithField(
			"error",
			err.Error(),
		).Error("failed to resolve item-recipes map")

		return err
	}

	logging.WithFields(logrus.Fields{
		"item-recipes":         len(req.ItemRecipesMap),
		"current-item-recipes": len(currentIrMap),
	}).Info("found current item-recipes")

	nextIrMap := currentIrMap.Merge(req.ItemRecipesMap)

	logging.WithFields(logrus.Fields{
		"item-recipes":         len(req.ItemRecipesMap),
		"current-item-recipes": len(currentIrMap),
		"merged-item-recipes":  len(nextIrMap),
	}).Info("resolved merged item-recipes")

	// pushing next ir-map out
	if err := sta.ProfessionsDatabase.PersistItemRecipes(req.Kind, nextIrMap); err != nil {
		logging.WithField("error", err.Error()).Error("failed to persist item-recipes")

		return err
	}

	logging.WithFields(logrus.Fields{
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("persisted item-recipes")

	return nil
}
