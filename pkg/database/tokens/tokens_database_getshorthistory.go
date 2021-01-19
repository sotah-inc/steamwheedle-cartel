package tokens

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type ShortTokenHistoryBatch map[blizzardv2.RegionName]TokenHistory

func NewShortTokenHistory(batch ShortTokenHistoryBatch) ShortTokenHistory {
	// gathering region-names
	regionNames := make([]blizzardv2.RegionName, len(batch))
	i := 0
	for regionName := range batch {
		regionNames[i] = regionName

		i += 1
	}

	// gathering a blank short-token-history
	out := ShortTokenHistory{}
	for _, tokenHistory := range batch {
		for unixTimestamp := range tokenHistory {
			normalizedTimestamp := sotah.NormalizeToDay(unixTimestamp)
			if _, ok := out[normalizedTimestamp]; ok {
				continue
			}

			out[normalizedTimestamp] = NewShortTokenHistoryItem(regionNames)
		}
	}

	return out
}

type ShortTokenHistory map[sotah.UnixTimestamp]ShortTokenHistoryItem

func NewShortTokenHistoryItem(regionNames []blizzardv2.RegionName) ShortTokenHistoryItem {
	out := ShortTokenHistoryItem{}
	for _, regionName := range regionNames {
		out[regionName] = 0
	}

	return out
}

type ShortTokenHistoryItem map[blizzardv2.RegionName]int64

type ShortTokenHistoryResponse struct {
	History map[sotah.UnixTimestamp]ShortTokenHistoryItem `json:"history"`
}

func (tBase Database) GetShortTokenHistory(regionNames []blizzardv2.RegionName) (ShortTokenHistoryResponse, error) {
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
		return ShortTokenHistoryResponse{}, err
	}

	out := ShortTokenHistoryResponse{
		History: NewShortTokenHistory(batch),
	}

	return out, nil
}
