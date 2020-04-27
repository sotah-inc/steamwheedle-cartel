package database

import (
	"errors"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type PricelistHistoryDatabases map[blizzardv2.RegionName]map[blizzardv2.ConnectedRealmId]PricelistHistoryDatabaseShards

func (phdBases PricelistHistoryDatabases) GetDatabase(
	tuple blizzardv2.LoadConnectedRealmTuple,
) (PricelistHistoryDatabase, error) {
	phdBase, ok := phdBases[tuple.RegionName][tuple.ConnectedRealmId][sotah.UnixTimestamp(tuple.LastModified.Unix())]
	if !ok {
		logging.WithFields(logrus.Fields{
			"region":          tuple.RegionName,
			"connected-realm": tuple.ConnectedRealmId,
			"last-modified":   tuple.LastModified.Unix(),
		}).Error("failed to find pricelist-history database")

		return PricelistHistoryDatabase{}, errors.New("failed to find pricelist-history database")
	}

	return phdBase, nil
}

type PricelistHistoryDatabaseShards map[sotah.UnixTimestamp]PricelistHistoryDatabase
