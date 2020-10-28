package areamaps

import (
	"strconv"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

// gathering area-maps
func (amBase Database) GetAreaMaps() (sotah.AreaMapMap, error) {
	out := sotah.AreaMapMap{}

	err := amBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName())
		if bkt == nil {
			return nil
		}

		err := bkt.ForEach(func(k, v []byte) error {
			parsedId, err := strconv.Atoi(string(k)[len("area-map-"):])
			if err != nil {
				return err
			}
			areaMapId := sotah.AreaMapId(parsedId)

			aMap, err := sotah.NewAreaMap(v)
			if err != nil {
				return err
			}

			out[areaMapId] = aMap

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return sotah.AreaMapMap{}, err
	}

	return out, nil
}

func (amBase Database) GetIdNormalizedNameMap() (sotah.AreaMapIdNameMap, error) {
	out := sotah.AreaMapIdNameMap{}

	err := amBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(namesBucketName())
		if bkt == nil {
			return nil
		}

		err := bkt.ForEach(func(k, v []byte) error {
			areaMapId, err := idFromNameKeyName(k)
			if err != nil {
				return err
			}

			out[areaMapId] = string(v)

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return sotah.AreaMapIdNameMap{}, err
	}

	return out, nil
}

func (amBase Database) FindAreaMaps(areaMapIds []sotah.AreaMapId) (sotah.AreaMapMap, error) {
	out := sotah.AreaMapMap{}
	err := amBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName())
		if bkt == nil {
			return nil
		}

		for _, id := range areaMapIds {
			value := bkt.Get(baseKeyName(id))
			if value == nil {
				continue
			}

			aMap, err := sotah.NewAreaMap(value)
			if err != nil {
				return err
			}

			out[id] = aMap
		}

		return nil
	})
	if err != nil {
		return sotah.AreaMapMap{}, err
	}

	return out, nil
}

// persisting
func (amBase Database) PersistAreaMaps(aMapMap sotah.AreaMapMap) error {
	logging.WithField("area-maps", len(aMapMap)).Debug("persisting area-maps")

	err := amBase.db.Batch(func(tx *bolt.Tx) error {
		areaMapsBucket, err := tx.CreateBucketIfNotExists(baseBucketName())
		if err != nil {
			return err
		}

		areaMapNamesBucket, err := tx.CreateBucketIfNotExists(namesBucketName())
		if err != nil {
			return err
		}

		for id, aMap := range aMapMap {
			jsonEncoded, err := aMap.EncodeForStorage()
			if err != nil {
				return err
			}

			if err := areaMapsBucket.Put(baseKeyName(id), jsonEncoded); err != nil {
				return err
			}

			normalizedName, err := sotah.NormalizeString(aMap.NormalizedName)
			if err != nil {
				return err
			}

			if err := areaMapNamesBucket.Put(nameKeyName(id), []byte(normalizedName)); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
