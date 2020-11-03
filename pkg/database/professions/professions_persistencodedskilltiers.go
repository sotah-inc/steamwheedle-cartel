package professions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

type PersistEncodedSkillTiersInJob struct {
	SkillTierId      blizzardv2.SkillTierId
	EncodedSkillTier []byte
}

func (pdBase Database) PersistEncodedSkillTiers(
	professionId blizzardv2.ProfessionId,
	in chan PersistEncodedSkillTiersInJob,
) (int, error) {
	logging.Info("persisting encoded skill-tiers")

	totalPersisted := 0

	err := pdBase.db.Batch(func(tx *bolt.Tx) error {
		stBucket, err := tx.CreateBucketIfNotExists(skillTiersBucketName(professionId))
		if err != nil {
			return err
		}

		for job := range in {
			if err := stBucket.Put(skillTiersKeyName(job.SkillTierId), job.EncodedSkillTier); err != nil {
				return err
			}

			totalPersisted += 1
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return totalPersisted, nil
}
