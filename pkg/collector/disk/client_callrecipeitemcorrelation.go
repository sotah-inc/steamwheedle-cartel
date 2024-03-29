package disk

import (
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemclass"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/professions/itemrecipekind" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

const RecipeItemClassId = itemclass.Recipe

func (c Client) CallRecipeItemCorrelation() error {
	startTime := time.Now()

	itemSubjectsByItemClassRequest := state.ItemSubjectsByItemClassRequest{
		ItemClassId: RecipeItemClassId,
		Version:     "",
	}
	encodedRequest, err := itemSubjectsByItemClassRequest.EncodeForDelivery()
	if err != nil {
		return err
	}

	// resolving item-subjects
	itemSubjectsMessage, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.ItemSubjectsByItemClass),
		Data:    encodedRequest,
		Timeout: 10 * time.Minute,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error(
			"failed to resolve item-subjects for item-class 9 (recipes)",
		)

		return err
	}

	if itemSubjectsMessage.Code != codes.Ok {
		logging.WithFields(
			itemSubjectsMessage.ToLogrusFields(),
		).Error("item-subjects request failed")

		return errors.New(itemSubjectsMessage.Err)
	}

	isMap, err := blizzardv2.NewItemSubjectsMap(itemSubjectsMessage.Data)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to decode item-subjects map")

		return err
	}

	logging.WithField("item-subjects", len(isMap)).Info("received item-subjects")

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

	irMap, err := blizzardv2.NewItemRecipesMapFromGzip(professionsMatchingItemsMessage.Data)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to decode item-recipes map")

		return err
	}

	logging.WithField("item-recipes-map", irMap).Info("received item-recipes")

	if professionsMatchingItemsMessage.Code != codes.Ok {
		logging.WithFields(
			itemSubjectsMessage.ToLogrusFields(),
		).Error("matching-items request failed")

		return errors.New(professionsMatchingItemsMessage.Err)
	}

	req := state.ItemRecipesIntakeRequest{
		Kind:           itemrecipekind.Teaches,
		ItemRecipesMap: irMap,
	}
	encodedItemRecipesIntakeRequest, err := req.EncodeForDelivery()
	if err != nil {
		logging.WithField(
			"error",
			err.Error(),
		).Error("failed to encode item-recipes-intake request")

		return err
	}

	itemRecipesIntakeMessage, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.ItemRecipesIntake),
		Data:    []byte(encodedItemRecipesIntakeRequest),
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
