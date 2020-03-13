package database

import (
	"strconv"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func NewAreaMapsDatabase(dbDir string) (AreaMapsDatabase, error) {
	dbFilepath, err := AreaMapsDatabasePath(dbDir)
	if err != nil {
		return AreaMapsDatabase{}, err
	}

	logging.WithField("filepath", dbFilepath).Info("initializing area-maps database")

	db, err := bolt.Open(dbFilepath, 0600, nil)
	if err != nil {
		return AreaMapsDatabase{}, err
	}

	return AreaMapsDatabase{db}, nil
}

type AreaMapsDatabase struct {
	db *bolt.DB
}

// gathering area-maps
func (amBase AreaMapsDatabase) GetAreaMaps() (sotah.AreaMapMap, error) {
	out := sotah.AreaMapMap{}

	err := amBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(databaseAreaMapsBucketName())
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

func (amBase AreaMapsDatabase) GetIdNormalizedNameMap() (sotah.AreaMapIdNameMap, error) {
	out := sotah.AreaMapIdNameMap{}

	err := amBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(databaseAreaMapNamesBucketName())
		if bkt == nil {
			return nil
		}

		err := bkt.ForEach(func(k, v []byte) error {
			areaMapId, err := areaMapIdFromAreaMapNameKeyName(k)
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

func (amBase AreaMapsDatabase) FindAreaMaps(areaMapIds []sotah.AreaMapId) (sotah.AreaMapMap, error) {
	out := sotah.AreaMapMap{}
	err := amBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(databaseAreaMapsBucketName())
		if bkt == nil {
			return nil
		}

		for _, id := range areaMapIds {
			value := bkt.Get(areaMapsKeyName(id))
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
func (amBase AreaMapsDatabase) PersistAreaMaps(aMapMap sotah.AreaMapMap) error {
	logging.WithField("area-maps", len(aMapMap)).Debug("Persisting area-maps")

	err := amBase.db.Batch(func(tx *bolt.Tx) error {
		areaMapsBucket, err := tx.CreateBucketIfNotExists(databaseAreaMapsBucketName())
		if err != nil {
			return err
		}

		areaMapNamesBucket, err := tx.CreateBucketIfNotExists(databaseAreaMapNamesBucketName())
		if err != nil {
			return err
		}

		for id, aMap := range aMapMap {
			jsonEncoded, err := aMap.EncodeForStorage()
			if err != nil {
				return err
			}

			if err := areaMapsBucket.Put(areaMapsKeyName(id), jsonEncoded); err != nil {
				return err
			}

			normalizedName, err := sotah.NormalizeName(aMap.NormalizedName)
			if err != nil {
				return err
			}

			if err := areaMapNamesBucket.Put(areaMapNameKeyName(id), []byte(normalizedName)); err != nil {
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
