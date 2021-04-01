package disk

import (
	"errors"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (c Client) CallEnchantingRecipeCorrelation() error {
	// resolving recipe-names
	recipeDescriptionMessage, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.ProfessionRecipeDescriptions),
		Data:    []byte(strconv.Itoa(int(blizzardv2.ProfessionId(333)))),
		Timeout: 10 * time.Minute,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error(
			"failed to resolve profession-recipe-names for profession 202 (enchanting)",
		)

		return err
	}

	if recipeDescriptionMessage.Code != codes.Ok {
		logging.WithFields(
			recipeDescriptionMessage.ToLogrusFields(),
		).Error("profession-recipe-names request failed")

		return errors.New(recipeDescriptionMessage.Err)
	}

	rdMap, err := blizzardv2.NewRecipeIdDescriptionMap(recipeDescriptionMessage.Data)
	if err != nil {
		logging.WithField(
			"error",
			err.Error(),
		).Error("failed to decode response data for recipe-name map")

		return err
	}

	logging.WithField("recipe-descriptions", rdMap).Info("found recipe-descriptions")

	encodedRdMap, err := rdMap.EncodeForDelivery()
	if err != nil {
		logging.WithField(
			"error",
			err.Error(),
		).Error("failed to encode recipe-name map")

		return err
	}

	// resolving matching items
	matchingItemsMessage, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.ItemsFindMatchingRecipes),
		Data:    []byte(encodedRdMap),
		Timeout: 10 * time.Minute,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error(
			"failed to resolve items-find-matching-recipes",
		)

		return err
	}

	if matchingItemsMessage.Code != codes.Ok {
		logging.WithFields(
			matchingItemsMessage.ToLogrusFields(),
		).Error("items-find-matching-recipes request failed")

		return errors.New(matchingItemsMessage.Err)
	}

	matchingItems, err := blizzardv2.NewItemRecipesMap([]byte(matchingItemsMessage.Data))
	if err != nil {
		logging.WithField(
			"error",
			err.Error(),
		).Error("failed to decode response data for matching-items")

		return err
	}

	logging.WithFields(logrus.Fields{
		"recipe-names":   rdMap,
		"matching-items": matchingItems,
	}).Info("found matches")

	return nil
}
