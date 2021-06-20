package items

import (
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
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

func (idBase Database) GetItemIds(version gameversion.GameVersion) (blizzardv2.ItemIds, error) {
	out := blizzardv2.ItemIds{}

	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName(version))
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

func (idBase Database) GetIdNormalizedNameMap(
	version gameversion.GameVersion,
) (sotah.ItemIdNameMap, error) {
	out := sotah.ItemIdNameMap{}

	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(namesBucketName(version))
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

type FindItemsJob struct {
	Err    error
	Id     blizzardv2.ItemId
	Item   sotah.Item
	Exists bool
}

func (job FindItemsJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"item":  job.Id,
	}
}

func (idBase Database) FindItems(
	version gameversion.GameVersion,
	ids blizzardv2.ItemIds,
) chan FindItemsJob {
	// starting up workers for gathering items
	in := make(chan blizzardv2.ItemId)
	out := make(chan FindItemsJob)
	worker := func() {
		for id := range in {
			item, exists, err := idBase.GetItem(version, id)
			if err != nil {
				out <- FindItemsJob{
					Err:    err,
					Id:     id,
					Item:   sotah.Item{},
					Exists: false,
				}

				continue
			}

			out <- FindItemsJob{
				Err:    nil,
				Id:     id,
				Item:   item,
				Exists: exists,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	// queueing it up
	go func() {
		for _, id := range ids {
			in <- id
		}

		close(in)
	}()

	return out
}

func (idBase Database) GetItem(
	version gameversion.GameVersion,
	id blizzardv2.ItemId,
) (sotah.Item, bool, error) {
	out := sotah.Item{}
	exists := false

	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName(version))
		if bkt == nil {
			return nil
		}

		v := bkt.Get(baseKeyName(id))
		if v == nil {
			return nil
		}

		exists = true

		item, err := sotah.NewItemFromGzipped(v)
		if err != nil {
			return err
		}

		out = item

		return nil
	})
	if err != nil {
		return sotah.Item{}, false, err
	}

	return out, exists, nil
}

func (idBase Database) ResetItems(version gameversion.GameVersion) error {
	bucketNames := [][]byte{
		baseBucketName(version),
		namesBucketName(version),
		blacklistBucketName(version),
		itemClassItemsBucket(),
	}

	for _, bucketName := range bucketNames {
		err := idBase.db.Batch(func(tx *bolt.Tx) error {
			bkt := tx.Bucket(bucketName)
			if bkt == nil {
				return nil
			}

			return bkt.ForEach(func(k []byte, v []byte) error {
				return bkt.Delete(k)
			})
		})
		if err != nil {
			return err
		}
	}

	return nil
}
