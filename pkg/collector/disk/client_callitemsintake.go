package disk

import (
	"errors"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (c Client) CallItemsIntake(
	viMap blizzardv2.VersionItemsMap,
) (state.ItemsIntakeResponse, error) {
	if viMap.IsZero() {
		return state.ItemsIntakeResponse{}, nil
	}

	// forwarding the received version-items map to items-intake
	encodedMap, err := viMap.EncodeForDelivery()
	if err != nil {
		logging.WithField(
			"error",
			err.Error(),
		).Error("failed to encode version-items map for delivery")

		return state.ItemsIntakeResponse{}, err
	}

	response, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.ItemsIntake),
		Data:    encodedMap,
		Timeout: 10 * time.Minute,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error(
			"failed to publish message for item-ids intake",
		)

		return state.ItemsIntakeResponse{}, err
	}

	if response.Code != codes.Ok {
		logging.WithFields(response.ToLogrusFields()).Error("item-ids intake request failed")

		return state.ItemsIntakeResponse{}, errors.New(response.Err)
	}

	itemsIntakeResponse, err := state.NewItemsIntakeResponse([]byte(response.Data))
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to parse item-intake response")

		return state.ItemsIntakeResponse{}, err
	}

	return itemsIntakeResponse, nil
}
