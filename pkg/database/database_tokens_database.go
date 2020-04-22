package database

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

func NewTokensDatabase(dbDir string) (TokensDatabase, error) {
	dbFilepath, err := TokensDatabasePath(dbDir)
	if err != nil {
		return TokensDatabase{}, err
	}

	logging.WithField("filepath", dbFilepath).Info("initializing tokens database")

	db, err := bolt.Open(dbFilepath, 0600, nil)
	if err != nil {
		return TokensDatabase{}, err
	}

	return TokensDatabase{db}, nil
}

type TokensDatabase struct {
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

// gathering token history
func (tBase TokensDatabase) GetHistory(regionName blizzardv2.RegionName) (TokenHistory, error) {
	out := TokenHistory{}

	err := tBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(databaseTokensBucketName(regionName))
		if bkt == nil {
			return nil
		}

		err := bkt.ForEach(func(k, v []byte) error {
			lastUpdated, err := lastUpdatedFromTokenKeyName(k)
			if err != nil {
				return err
			}

			out[lastUpdated] = priceFromTokenValue(v)

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return TokenHistory{}, err
	}

	return out, nil
}

// persisting
func (tBase TokensDatabase) PersistHistory(rtHistory RegionTokenHistory) error {
	logging.WithField("region-token-history", rtHistory).Debug("persisting region token-history")

	err := tBase.db.Batch(func(tx *bolt.Tx) error {
		for regionName, tHistory := range rtHistory {
			bkt, err := tx.CreateBucketIfNotExists(databaseTokensBucketName(regionName))
			if err != nil {
				return err
			}

			for lastUpdated, price := range tHistory {
				if err := bkt.Put(tokenKeyName(lastUpdated), priceToTokenValue(price)); err != nil {
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
func (tBase TokensDatabase) Prune(regionNames []blizzardv2.RegionName) error {
	earliestUnixTimestamp := RetentionLimit().Unix()

	err := tBase.db.Update(func(tx *bolt.Tx) error {
		for _, regionName := range regionNames {
			bkt := tx.Bucket(databaseTokensBucketName(regionName))
			if bkt == nil {
				continue
			}

			c := bkt.Cursor()

			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				lastUpdated, err := lastUpdatedFromTokenKeyName(k)
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
