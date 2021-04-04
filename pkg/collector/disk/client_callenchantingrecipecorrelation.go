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
	startTime := time.Now()

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

	// resolving matching items
	matchingItemsMessage, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.ItemsFindMatchingRecipes),
		Data:    []byte(recipeDescriptionMessage.Data),
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

	itemRecipesIntakeMessage, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.ItemRecipesIntake),
		Data:    []byte(matchingItemsMessage.Data),
		Timeout: 10 * time.Minute,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error(
			"failed to publish message for item-recipes intake",
		)

		return err
	}

	if itemRecipesIntakeMessage.Code != codes.Ok {
		logging.WithFields(
			itemRecipesIntakeMessage.ToLogrusFields(),
		).Error("item-recipes intake request failed")

		return err
	}

	logging.WithFields(logrus.Fields{
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("finished enchanting-recipe-correlation")

	return nil
}
