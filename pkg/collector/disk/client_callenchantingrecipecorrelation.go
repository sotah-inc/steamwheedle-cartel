package disk

import (
	"errors"
	"strconv"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (c Client) CallEnchantingRecipeCorrelation() error {
	recipeNamesMessage, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.ProfessionRecipeNames),
		Data:    []byte(strconv.Itoa(int(blizzardv2.ProfessionId(333)))),
		Timeout: 10 * time.Minute,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error(
			"failed to resolve profession-recipe-names for profession 202 (enchanting)",
		)

		return err
	}

	if recipeNamesMessage.Code != codes.Ok {
		logging.WithFields(
			recipeNamesMessage.ToLogrusFields(),
		).Error("profession-recipe-names request failed")

		return errors.New(recipeNamesMessage.Err)
	}

	recipeNameMap, err := blizzardv2.NewRecipeIdNameMap(recipeNamesMessage.Data)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to decode response data")

		return err
	}

	logging.WithField("recipe-name-map", recipeNameMap).Info("received data")

	return nil
}
