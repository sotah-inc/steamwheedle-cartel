package tokens

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type ShortTokenHistoryBatch map[blizzardv2.RegionName]TokenHistory

func NewShortTokenHistory(batch ShortTokenHistoryBatch) ShortTokenHistory {
	// resolving batched batches
	for regionName, tHistory := range batch {
		batch[regionName] = NewTokenHistoryFromBatch(
			NewTokenHistoryBatch(
				tHistory,
				func(targetTimestamp sotah.UnixTimestamp) sotah.UnixTimestamp {
					return sotah.NormalizeToHour(targetTimestamp/1000) * 1000
				},
			),
		)
	}

	// gathering region-names
	regionNames := make([]blizzardv2.RegionName, len(batch))
	i := 0
	for regionName := range batch {
		regionNames[i] = regionName

		i += 1
	}

	// gathering timestamps
	timestamps := func() sotah.UnixTimestamps {
		foundTimestampsMap := map[sotah.UnixTimestamp]struct{}{}
		for _, tHistory := range batch {
			for timestamp := range tHistory {
				foundTimestampsMap[timestamp] = struct{}{}
			}
		}

		foundTimestamps := make(sotah.UnixTimestamps, len(foundTimestampsMap))
		i := 0
		for timestamp := range foundTimestampsMap {
			foundTimestamps[i] = timestamp

			i += 1
		}

		return foundTimestamps
	}()

	// merging results together
	out := ShortTokenHistory{}
	for _, timestamp := range timestamps {
		out[timestamp] = func() ShortTokenHistoryItem {
			item := ShortTokenHistoryItem{}
			for _, regionName := range regionNames {
				item[regionName] = func() int64 {
					foundRegionBatch, ok := batch[regionName]
					if !ok {
						return 0
					}

					foundPrice, ok := foundRegionBatch[timestamp]
					if !ok {
						return 0
					}

					return foundPrice
				}()
			}

			return item
		}()

	}

	return out
}

type ShortTokenHistory map[sotah.UnixTimestamp]ShortTokenHistoryItem

type ShortTokenHistoryItem map[blizzardv2.RegionName]int64

func (tBase Database) GetShortTokenHistory(
	regionNames []blizzardv2.RegionName,
) (ShortTokenHistory, error) {
	batch := ShortTokenHistoryBatch{}
	err := tBase.db.View(func(tx *bolt.Tx) error {
		for _, regionName := range regionNames {
			batch[regionName] = TokenHistory{}

			bkt := tx.Bucket(baseBucketName(regionName))
			if bkt == nil {
				return nil
			}

			err := bkt.ForEach(func(k, v []byte) error {
				lastUpdated, err := lastUpdatedFromBaseKeyName(k)
				if err != nil {
					return err
				}

				batch[regionName][lastUpdated] = priceFromTokenValue(v)

				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return ShortTokenHistory{}, err
	}

	return NewShortTokenHistory(batch), nil
}
