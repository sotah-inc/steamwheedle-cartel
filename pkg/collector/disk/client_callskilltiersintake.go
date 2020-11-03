package disk

import (
	"errors"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (c Client) CallSkillTiersIntake() error {
	professionsMsg, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.Professions),
		Data:    nil,
		Timeout: 10 * time.Minute,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error(
			"failed to publish message for get-professions",
		)

		return err
	}

	if professionsMsg.Code != codes.Ok {
		logging.WithFields(professionsMsg.ToLogrusFields()).Error("get-professions request failed")

		return errors.New(professionsMsg.Err)
	}

	professionsResponse, err := state.NewProfessionsResponse(professionsMsg.Data)
	if err != nil {
		logging.WithField("error", err.Error()).Error(
			"failed to parse response from gathering professions",
		)

		return err
	}

	for _, profession := range professionsResponse.Professions {
		req := sotah.SkillTiersIntakeRequest{ProfessionId: profession.BlizzardMeta.Id}
		encodedReq, err := req.EncodeForDelivery()
		if err != nil {
			logging.WithField("error", err.Error()).Error(
				"failed to encode skill-tiers-intake request",
			)

			return err
		}

		response, err := c.messengerClient.Request(messenger.RequestOptions{
			Subject: string(subjects.SkillTiersIntake),
			Timeout: 10 * time.Minute,
			Data:    encodedReq,
		})
		if err != nil {
			logging.WithField("error", err.Error()).Error(
				"failed to publish message for skill-tiers intake",
			)

			return err
		}

		if response.Code != codes.Ok {
			logging.WithFields(response.ToLogrusFields()).Error("skill-tiers intake request failed")

			return errors.New(response.Err)
		}
	}

	return nil
}
