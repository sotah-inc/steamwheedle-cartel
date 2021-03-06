package disk

import (
	"strconv"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type getEncodedSkillTierJob struct {
	err              error
	id               blizzardv2.SkillTierId
	encodedSkillTier []byte
}

func (g getEncodedSkillTierJob) Err() error                 { return g.err }
func (g getEncodedSkillTierJob) Id() blizzardv2.SkillTierId { return g.id }
func (g getEncodedSkillTierJob) EncodedSkillTier() []byte   { return g.encodedSkillTier }
func (g getEncodedSkillTierJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": g.err.Error(),
		"id":    g.id,
	}
}

func (client Client) GetEncodedSkillTiers(
	professionId blizzardv2.ProfessionId,
	idList []blizzardv2.SkillTierId,
) chan BaseLake.GetEncodedSkillTierJob {
	out := make(chan BaseLake.GetEncodedSkillTierJob)

	// starting up workers for resolving skill-tiers
	skillTiersOut := client.resolveSkillTiers(professionId, idList)

	logging.WithField(
		"primary-skill-tiers",
		client.primarySkillTiers,
	).Info("checking with skill-tiers")

	// queueing it all up
	go func() {
		for job := range skillTiersOut {
			if job.Err != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to resolve skill-tier")

				continue
			}

			isPrimary := func() bool {
				foundSkillTierPrimaries, ok := client.primarySkillTiers[strconv.FormatInt(
					int64(professionId),
					10,
				)]
				if !ok {
					return false
				}

				for _, id := range foundSkillTierPrimaries {
					if id == job.Id {
						return true
					}
				}

				return false
			}()

			skillTier := sotah.SkillTier{
				BlizzardMeta: job.SkillTierResponse,
				SotahMeta: sotah.SkillTierMeta{
					IsPrimary: isPrimary,
				},
			}

			encodedSkillTier, err := skillTier.EncodeForStorage()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":      err.Error(),
					"skill-tier": skillTier.BlizzardMeta.Id,
				}).Error("failed to encode skill-tier for storage")

				continue
			}

			out <- getEncodedSkillTierJob{
				err:              nil,
				id:               job.SkillTierResponse.Id,
				encodedSkillTier: encodedSkillTier,
			}
		}

		close(out)
	}()

	return out
}
