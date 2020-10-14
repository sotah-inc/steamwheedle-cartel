package disk

import (
	"errors"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (c Client) CallItemsIntake(ids blizzardv2.ItemIds) error {
	// forwarding the received item-ids to items-history intake
	encodedTuples, err := ids.EncodeForDelivery()
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to encode item-ids for delivery")

		return err
	}

	response, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.ItemsIntake),
		Data:    encodedTuples,
		Timeout: 10 * time.Minute,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error(
			"failed to publish message for item-ids intake",
		)

		return err
	}

	if response.Code != codes.Ok {
		logging.WithFields(response.ToLogrusFields()).Error("item-ids intake request failed")

		return errors.New(response.Err)
	}

	return nil
}
