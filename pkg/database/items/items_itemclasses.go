package items

import (
	"encoding/json"

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

func (idBase Database) PersistItemClasses(itemClasses []blizzardv2.ItemClassResponse) error {
	return idBase.db.Batch(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(itemClassesBucket())
		if err != nil {
			return err
		}

		encodedItemClasses, err := json.Marshal(itemClasses)
		if err != nil {
			return err
		}

		return bkt.Put(itemClassesKeyName(), encodedItemClasses)
	})
}
