package disk

import (
	"errors"
	"strconv"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

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

	// resolving item-subjects
	itemSubjectsMessage, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.ItemSubjectsByItemClass),
		Data:    []byte(strconv.Itoa(int(RecipeItemClassId))),
		Timeout: 10 * time.Minute,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error(
			"failed to resolve item-subjects for item-class 9 (recipes)",
		)

		return err
	}

	isMap, err := blizzardv2.NewItemSubjectsMap(itemSubjectsMessage.Data)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to decode item-subjects map")

		return err
	}

	logging.WithField("item-subjects", isMap).Info("received item-subjects")

	if itemSubjectsMessage.Code != codes.Ok {
		logging.WithFields(
			itemSubjectsMessage.ToLogrusFields(),
		).Error("item-subjects request failed")

		return errors.New(itemSubjectsMessage.Err)
	}

	// resolving item-recipes from professions
	professionsMatchingItemsMessage, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.ProfessionsFindMatchingItems),
		Data:    []byte(itemSubjectsMessage.Data),
		Timeout: 10 * time.Minute,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error(
			"failed to resolve matching-items",
		)

		return err
	}

	irMap, err := blizzardv2.NewItemRecipesMap(professionsMatchingItemsMessage.Data)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to decode item-recipes map")

		return err
	}

	logging.WithField("item-recipes", irMap).Info("received item-recipes")

	if professionsMatchingItemsMessage.Code != codes.Ok {
		logging.WithFields(
			itemSubjectsMessage.ToLogrusFields(),
		).Error("matching-items request failed")

		return errors.New(professionsMatchingItemsMessage.Err)
	}

	itemRecipesIntakeMessage, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.ItemRecipesIntake),
		Data:    []byte(professionsMatchingItemsMessage.Data),
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
	}).Info("finished recipe-item-correlation")

	return nil
}
