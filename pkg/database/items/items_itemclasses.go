package items

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (idBase Database) HasItemClasses() (bool, error) {
	out := false

	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(itemClassesBucket())
		if bkt == nil {
			return nil
		}

		v := bkt.Get(itemClassesKeyName())
		if v == nil {
			return nil
		}

		out = true

		return nil
	})
	if err != nil {
		return false, err
	}

	return out, nil
}

func (idBase Database) GetItemClasses() ([]blizzardv2.ItemClassResponse, error) {
	var out []blizzardv2.ItemClassResponse

	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(itemClassesBucket())
		if bkt == nil {
			return nil
		}

		v := bkt.Get(itemClassesKeyName())
		if v == nil {
			return nil
		}

		return json.Unmarshal(v, &out)
	})
	if err != nil {
		return []blizzardv2.ItemClassResponse{}, err
	}

	return out, nil
}

func (idBase Database) PersistItemClasses(encodedItemClasses []byte) error {
	logging.Info("persisting encoded item-classes")

	return idBase.db.Batch(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(itemClassesBucket())
		if err != nil {
			return err
		}

		return bkt.Put(itemClassesKeyName(), encodedItemClasses)
	})
}
