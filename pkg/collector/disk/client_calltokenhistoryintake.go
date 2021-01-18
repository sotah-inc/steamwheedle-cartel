package disk

import (
	"errors"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (c Client) CallTokenHistoryIntake() error {
	response, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.TokenHistoryIntake),
		Data:    nil,
		Timeout: 10 * time.Minute,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error(
			"failed to publish message for token-history intake",
		)

		return err
	}

	if response.Code != codes.Ok {
		logging.WithFields(response.ToLogrusFields()).Error("token-history intake request failed")

		return errors.New(response.Err)
	}

	return nil
}
