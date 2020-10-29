package professions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

type PersistEncodedProfessionsInJob struct {
	ProfessionId      blizzardv2.ProfessionId
	EncodedProfession []byte
}

func (pdBase Database) PersistEncodedProfessions(
	in chan PersistEncodedProfessionsInJob,
) (int, error) {
	logging.Info("persisting encoded professions")

	totalPersisted := 0

	err := pdBase.db.Batch(func(tx *bolt.Tx) error {
		baseBucket, err := tx.CreateBucketIfNotExists(baseBucketName())
		if err != nil {
			return err
		}

		for job := range in {
			if err := baseBucket.Put(baseKeyName(job.ProfessionId), job.EncodedProfession); err != nil {
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
