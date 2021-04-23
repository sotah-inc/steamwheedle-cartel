package disk

import (
	"errors"
	"strconv"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemclass"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

const RecipeItemClassId = itemclass.Recipe

func (c Client) CallRecipeItemCorrelation() error {
	startTime := time.Now()

	// resolving recipe-subjects
	recipeSubjectsMessage, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.ProfessionRecipeSubjects),
		Data:    []byte(strconv.Itoa(int(EnchantingProfessionId))),
		Timeout: 10 * time.Minute,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error(
			"failed to resolve profession-recipe-subjects for profession 202 (enchanting)",
		)

		return err
	}

	if recipeSubjectsMessage.Code != codes.Ok {
		logging.WithFields(
			recipeSubjectsMessage.ToLogrusFields(),
		).Error("profession-recipe-subjects request failed")

		return errors.New(recipeSubjectsMessage.Err)
	}

	// resolving matching items
	matchingItemsMessage, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.ItemsFindMatchingRecipes),
		Data:    []byte(recipeSubjectsMessage.Data),
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
