package items

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func NewDatabase(dbDir string) (Database, error) {
	dbFilepath, err := DatabasePath(dbDir)
	if err != nil {
		return Database{}, err
	}

	logging.WithField("filepath", dbFilepath).Info("initializing items database")

	db, err := bolt.Open(dbFilepath, 0600, nil)
	if err != nil {
		return Database{}, err
	}

	return Database{db}, nil
}

type Database struct {
	db *bolt.DB
}

// gathering items
func (idBase Database) GetItemIds() (blizzardv2.ItemIds, error) {
	out := blizzardv2.ItemIds{}

	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName())
		if bkt == nil {
			return nil
		}

		err := bkt.ForEach(func(k, v []byte) error {
			id, err := itemIdFromKeyName(k)
			if err != nil {
				return err
			}

			out = append(out, id)

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return blizzardv2.ItemIds{}, err
	}

	return out, nil
}

func (idBase Database) GetItems() (sotah.ItemsMap, error) {
	out := sotah.ItemsMap{}

	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName())
		if bkt == nil {
			return nil
		}

		err := bkt.ForEach(func(k, v []byte) error {
			id, err := itemIdFromKeyName(k)
			if err != nil {
				return err
			}

			out[id], err = sotah.NewItemFromGzipped(v)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return sotah.ItemsMap{}, err
	}

	return out, nil
}

func (idBase Database) GetIdNormalizedNameMap() (sotah.ItemIdNameMap, error) {
	out := sotah.ItemIdNameMap{}

	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(namesBucketName())
		if bkt == nil {
			return nil
		}

		err := bkt.ForEach(func(k, v []byte) error {
			itemId, err := itemIdFromNameKeyName(k)
			if err != nil {
				return err
			}

			out[itemId], err = locale.NewMapping(v)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return sotah.ItemIdNameMap{}, err
	}

	return out, nil
}

func (idBase Database) FindItems(itemIds blizzardv2.ItemIds) (sotah.ItemsMap, error) {
	out := sotah.ItemsMap{}
	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName())
		if bkt == nil {
			return nil
		}

		for _, id := range itemIds {
			value := bkt.Get(baseKeyName(id))
			if value == nil {
				continue
			}

			var err error
			out[id], err = sotah.NewItemFromGzipped(value)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return sotah.ItemsMap{}, err
	}

	return out, nil
}
