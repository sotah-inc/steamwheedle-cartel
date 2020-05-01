package database

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (shards PricelistHistoryDatabaseShards) GetPriceHistory(
	id blizzardv2.ItemId,
	lowerBounds sotah.UnixTimestamp,
	upperBounds sotah.UnixTimestamp,
) (sotah.PriceHistory, error) {
	pHistory := sotah.PriceHistory{}

	for _, phdBase := range shards {
		receivedHistory, err := phdBase.getItemPriceHistory(id)
		if err != nil {
			return sotah.PriceHistory{}, err
		}

		pHistory = pHistory.Merge(receivedHistory.Between(lowerBounds, upperBounds))
	}

	return pHistory, nil
}
