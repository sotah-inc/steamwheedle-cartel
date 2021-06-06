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

func (c Client) CallLiveAuctionsIntake(req state.IntakeRequest) error {
	// forwarding the received request to live-auctions intake
	encodedRequest, err := req.EncodeForDelivery()
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to encode load tuples for delivery")

		return err
	}

	response, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.LiveAuctionsIntake),
		Data:    encodedRequest,
		Timeout: 10 * time.Minute,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error(
			"failed to publish message for live-auctions intake",
		)

		return err
	}

	if response.Code != codes.Ok {
		logging.WithFields(response.ToLogrusFields()).Error("live-auctions intake request failed")

		return errors.New(response.Err)
	}

	return nil
}
