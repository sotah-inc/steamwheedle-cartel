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

	if itemSubjectsMessage.Code != codes.Ok {
		logging.WithFields(
			itemSubjectsMessage.ToLogrusFields(),
		).Error("item-subjects request failed")

		return errors.New(itemSubjectsMessage.Err)
	}

	// resolving item-recipes from professions
	professionsMatchingItemsMessage, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.ProfessionsFindMatchingItems),
		Data:    []byte(strconv.Itoa(int(RecipeItemClassId))),
		Timeout: 10 * time.Minute,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error(
			"failed to resolve matching-items",
		)

		return err
	}

	if professionsMatchingItemsMessage.Code != codes.Ok {
		logging.WithFields(
			itemSubjectsMessage.ToLogrusFields(),
		).Error("matching-items request failed")

		return errors.New(professionsMatchingItemsMessage.Err)
	}

	logging.WithFields(logrus.Fields{
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("finished recipe-item-correlation")

	return nil
}
