package tokens

import (
	"encoding/json"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

func NewDatabase(dbDir string) (Database, error) {
	dbFilepath, err := DatabasePath(dbDir)
	if err != nil {
		return Database{}, err
	}

	logging.WithField("filepath", dbFilepath).Info("initializing tokens database")

	db, err := bolt.Open(dbFilepath, 0600, nil)
	if err != nil {
		return Database{}, err
	}

	return Database{db}, nil
}

type Database struct {
	db *bolt.DB
}

type RegionTokenHistory map[blizzardv2.RegionName]TokenHistory

type TokenHistory map[int64]int64

func (tHistory TokenHistory) EncodeForDelivery() ([]byte, error) {
	jsonEncoded, err := json.Marshal(tHistory)
	if err != nil {
		return []byte{}, err
	}

	return jsonEncoded, nil
}

// persisting
func (tBase Database) PersistHistory(rtHistory RegionTokenHistory) error {
	logging.WithField("region-token-history", rtHistory).Debug("persisting region token-history")

	err := tBase.db.Batch(func(tx *bolt.Tx) error {
		for regionName, tHistory := range rtHistory {
			bkt, err := tx.CreateBucketIfNotExists(baseBucketName(regionName))
			if err != nil {
				return err
			}

			for lastUpdated, price := range tHistory {
				if err := bkt.Put(baseKeyName(lastUpdated), priceToTokenValue(price)); err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// pruning
func (tBase Database) Prune(regionNames []blizzardv2.RegionName) error {
	earliestUnixTimestamp := BaseDatabase.RetentionLimit().Unix()

	err := tBase.db.Update(func(tx *bolt.Tx) error {
		for _, regionName := range regionNames {
			bkt := tx.Bucket(baseBucketName(regionName))
			if bkt == nil {
				continue
			}

			c := bkt.Cursor()

			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				lastUpdated, err := lastUpdatedFromBaseKeyName(k)
				if err != nil {
					return err
				}

				if lastUpdated > earliestUnixTimestamp {
					continue
				}

				if err := bkt.Delete(k); err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
