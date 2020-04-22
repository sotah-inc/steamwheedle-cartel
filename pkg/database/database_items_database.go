package database

import (
	"strconv"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func NewItemsDatabase(dbDir string) (ItemsDatabase, error) {
	dbFilepath, err := ItemsDatabasePath(dbDir)
	if err != nil {
		return ItemsDatabase{}, err
	}

	logging.WithField("filepath", dbFilepath).Info("initializing items database")

	db, err := bolt.Open(dbFilepath, 0600, nil)
	if err != nil {
		return ItemsDatabase{}, err
	}

	return ItemsDatabase{db}, nil
}

type ItemsDatabase struct {
	db *bolt.DB
}

// gathering items
func (idBase ItemsDatabase) GetItems() (sotah.ItemsMap, error) {
	out := sotah.ItemsMap{}

	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(databaseItemsBucketName())
		if bkt == nil {
			return nil
		}

		err := bkt.ForEach(func(k, v []byte) error {
			parsedId, err := strconv.Atoi(string(k)[len("item-"):])
			if err != nil {
				return err
			}
			itemId := blizzardv2.ItemId(parsedId)

			out[itemId], err = sotah.NewItemFromGzipped(v)
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

func (idBase ItemsDatabase) GetIdNormalizedNameMap() (sotah.ItemIdNameMap, error) {
	out := sotah.ItemIdNameMap{}

	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(databaseItemNamesBucketName())
		if bkt == nil {
			return nil
		}

		err := bkt.ForEach(func(k, v []byte) error {
			itemId, err := itemIdFromItemNameKeyName(k)
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

func (idBase ItemsDatabase) FindItems(itemIds []blizzardv2.ItemId) (sotah.ItemsMap, error) {
	out := sotah.ItemsMap{}
	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(databaseItemsBucketName())
		if bkt == nil {
			return nil
		}

		for _, id := range itemIds {
			value := bkt.Get(itemsKeyName(id))
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
