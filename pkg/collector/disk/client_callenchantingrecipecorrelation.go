package disk

import (
	"encoding/base64"
	"errors"
	"strconv"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"

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

	data := []byte(recipeDescriptionMessage.Data)

	logging.WithField(
		"recipeDescriptionMessage.Data-length",
		len(data),
	).Info("received recipe-description response")

	// resolving matching items
	matchingItemsMessage, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.ItemsFindMatchingRecipes),
		Data:    data,
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

	gzipEncoded, err := base64.StdEncoding.DecodeString(matchingItemsMessage.Data)
	if err != nil {
		logging.WithField(
			"error",
			err.Error(),
		).Error("failed to base64-decode matching-items message")

		return err
	}

	jsonEncoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		logging.WithField(
			"error",
			err.Error(),
		).Error("failed to gzip-decode matching-items message")

		return err
	}

	matchingItems, err := blizzardv2.NewItemRecipesMap(jsonEncoded)
	if err != nil {
		logging.WithField(
			"error",
			err.Error(),
		).Error("failed to decode response data for matching-items")

		return err
	}

	logging.WithFields(logrus.Fields{
		"matching-items": matchingItems,
	}).Info("found matches")

	return nil
}
