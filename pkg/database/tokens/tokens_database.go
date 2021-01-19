package tokens

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"

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

func NewTokenHistoryFromBatch(batch TokenHistoryBatch) TokenHistory {
	out := TokenHistory{}
	for timestamp, prices := range batch {
		priceAverage := func() int64 {
			total := int64(0)
			for _, price := range prices {
				total += price
			}

			return total / int64(len(prices))
		}()

		out[timestamp] = priceAverage
	}

	return out
}

type TokenHistory map[sotah.UnixTimestamp]int64

func (tHistory TokenHistory) EncodeForDelivery() ([]byte, error) {
	jsonEncoded, err := json.Marshal(tHistory)
	if err != nil {
		return []byte{}, err
	}

	return jsonEncoded, nil
}

func NewTokenHistoryBatch(
	tHistory TokenHistory,
	normalizeFunc func(sotah.UnixTimestamp) sotah.UnixTimestamp,
) TokenHistoryBatch {
	out := TokenHistoryBatch{}
	for timestamp, price := range tHistory {
		normalizedTimestamp := normalizeFunc(timestamp)
		batch := func() []int64 {
			foundBatch, ok := out[normalizedTimestamp]
			if !ok {
				return []int64{}
			}

			return foundBatch
		}()

		batch = append(batch, price)
		out[normalizedTimestamp] = batch
	}

	return out
}

type TokenHistoryBatch map[sotah.UnixTimestamp][]int64

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
				if err := bkt.Put(baseKeyName(int64(lastUpdated)), priceToTokenValue(price)); err != nil {
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
	earliestUnixTimestamp := sotah.UnixTimestamp(BaseDatabase.RetentionLimit().Unix())

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
