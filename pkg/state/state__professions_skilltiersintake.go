package state

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	ProfessionsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/professions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta ProfessionsState) ListenForSkillTiersIntake(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.SkillTiersIntake), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		req, err := sotah.NewSkillTiersIntakeRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		if err := sta.SkillTiersIntake(req.ProfessionId); err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}

func (sta ProfessionsState) SkillTiersIntake(professionId blizzardv2.ProfessionId) error {
	startTime := time.Now()

	// resolving the profession
	profession, err := sta.ProfessionsDatabase.GetProfession(professionId)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    professionId,
		}).Error("failed to fetch profession")

		return err
	}

	skillTierIdsToFetch := make([]blizzardv2.SkillTierId, len(profession.BlizzardMeta.SkillTiers))
	for i, skillTier := range profession.BlizzardMeta.SkillTiers {
		skillTierIdsToFetch[i] = skillTier.Id
	}

	// resolving skill-tier-ids to not sync
	skillTierIdsBlacklist, err := sta.ProfessionsDatabase.GetSkillTierIds(professionId)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to get skill-tier-ids for blacklist")

		return err
	}

	logging.WithFields(logrus.Fields{
		"ids-to-fetch":  len(skillTierIdsToFetch),
		"ids-blacklist": len(skillTierIdsBlacklist),
	}).Info("collecting skill-tiers sans blacklist")

	// starting up an intake queue
	getEncodedSkillTiersOut := sta.LakeClient.GetEncodedSkillTiers(
		professionId,
		skillTierIdsToFetch,
		skillTierIdsBlacklist,
	)
	persistSkillTiersIn := make(chan ProfessionsDatabase.PersistEncodedSkillTiersInJob)

	// queueing it all up
	go func() {
		for job := range getEncodedSkillTiersOut {
			if job.Err() != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to resolve skill-tier")

				continue
			}

			logging.WithField("skill-tier-id", job.Id()).Info("enqueueing skill-tier for persistence")

			persistSkillTiersIn <- ProfessionsDatabase.PersistEncodedSkillTiersInJob{
				SkillTierId:      job.Id(),
				EncodedSkillTier: job.EncodedSkillTier(),
			}
		}

		close(persistSkillTiersIn)
	}()

	totalPersisted, err := sta.ProfessionsDatabase.PersistEncodedSkillTiers(
		profession.BlizzardMeta.Id,
		persistSkillTiersIn,
	)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to persist skill-tier")

		return err
	}

	logging.WithFields(logrus.Fields{
		"total":          totalPersisted,
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total persisted in skill-tier-intake")

	return nil
}
