package professions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/professions/professionsflags" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

func NewDatabase(dbDir string) (Database, error) {
	dbFilepath, err := DatabasePath(dbDir)
	if err != nil {
		return Database{}, err
	}

	logging.WithField("filepath", dbFilepath).Info("initializing professions database")

	db, err := bolt.Open(dbFilepath, 0600, nil)
	if err != nil {
		return Database{}, err
	}

	return Database{db}, nil
}

type Database struct {
	db *bolt.DB
}

func (pdBase Database) IsComplete(flag professionsflags.ProfessionFlag) (bool, error) {
	out := false

	err := pdBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(flagsBucketName())
		if bkt == nil {
			return nil
		}

		data := bkt.Get(isCompleteKeyName(flag))
		if data == nil {
			return nil
		}

		if string(data) == "1" {
			out = true
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	return out, nil
}

func (pdBase Database) IsAllComplete() (bool, error) {
	flags := []professionsflags.ProfessionFlag{
		professionsflags.Professions,
		professionsflags.Recipes,
		professionsflags.SkillTiers,
	}
	for _, flag := range flags {
		isComplete, err := pdBase.IsComplete(flag)
		if err != nil {
			return false, err
		}

		if !isComplete {
			return false, nil
		}
	}

	return true, nil
}

func (pdBase Database) SetIsComplete(flag professionsflags.ProfessionFlag) error {
	return pdBase.db.Batch(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(flagsBucketName())
		if err != nil {
			return err
		}

		if err := bkt.Put(isCompleteKeyName(flag), []byte("1")); err != nil {
			return err
		}

		return nil
	})
}
