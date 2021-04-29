package disk

import (
	"errors"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (c Client) CallRecipesIntake() (state.RecipesIntakeResponse, error) {
	response, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.RecipesIntake),
		Timeout: 10 * time.Minute,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error(
			"failed to publish message for recipes intake",
		)

		return state.RecipesIntakeResponse{}, err
	}

	if response.Code != codes.Ok {
		logging.WithFields(response.ToLogrusFields()).Error("recipes intake request failed")

		return state.RecipesIntakeResponse{}, errors.New(response.Err)
	}

	recipesIntakeResponse, err := state.NewRecipesIntakeResponse(response.Data)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to parse recipes-intake response")

		return state.RecipesIntakeResponse{}, errors.New(response.Err)
	}

	return recipesIntakeResponse, nil
}
