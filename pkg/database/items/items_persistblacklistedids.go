package items

import (
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

func (idBase Database) PersistBlacklistedIds(
	version gameversion.GameVersion,
	ids []blizzardv2.ItemId,
) error {
	logging.WithField("erroneous-ids", ids).Info("persisting blacklisted item-ids")

	return idBase.db.Batch(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(blacklistBucketName())
		if err != nil {
			return err
		}

		for _, id := range ids {
			tuple := blizzardv2.VersionItemTuple{
				GameVersion: version,
				Id:          id,
			}

			logging.WithFields(logrus.Fields{
				"tuple": tuple,
			}).Info("persisting blacklisted item-id")

			if err := bkt.Put(blacklistKeyName(tuple), blacklistKeyName(tuple)); err != nil {
				return err
			}
		}

		return nil
	})
}
