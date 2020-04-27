package database

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func NewPricelistHistoryDatabases(
	dirPath string,
	tuples blizzardv2.RegionConnectedRealmTuples,
) (*PricelistHistoryDatabases, error) {
	if dirPath == "" {
		return nil, errors.New("dir-path cannot be blank")
	}

	phdBases := PricelistHistoryDatabases{
		databaseDir: dirPath,
		Databases:   map[blizzardv2.RegionName]map[blizzardv2.ConnectedRealmId]PricelistHistoryDatabaseShards{},
	}

	for _, tuple := range tuples {
		if _, ok := phdBases.Databases[tuple.RegionName]; !ok {
			phdBases.Databases[tuple.RegionName] = map[blizzardv2.ConnectedRealmId]PricelistHistoryDatabaseShards{}
		}
		if _, ok := phdBases.Databases[tuple.RegionName][tuple.ConnectedRealmId]; !ok {
			phdBases.Databases[tuple.RegionName][tuple.ConnectedRealmId] = PricelistHistoryDatabaseShards{}
		}

		dbPathPairs, err := Paths(fmt.Sprintf(
			"%s/pricelist-histories/%s/%d",
			dirPath,
			tuple.RegionName,
			tuple.ConnectedRealmId,
		))
		if err != nil {
			return nil, err
		}

		for _, dbPathPair := range dbPathPairs {
			phdBase, err := newPricelistHistoryDatabase(dbPathPair.FullPath, dbPathPair.Timestamp)
			if err != nil {
				return nil, err
			}

			phdBases.Databases[tuple.RegionName][tuple.ConnectedRealmId][dbPathPair.Timestamp] = phdBase
		}
	}

	return &phdBases, nil
}

type PricelistHistoryDatabases struct {
	databaseDir string
	Databases   map[blizzardv2.RegionName]map[blizzardv2.ConnectedRealmId]PricelistHistoryDatabaseShards
}

func (phdBases *PricelistHistoryDatabases) GetDatabase(
	tuple blizzardv2.LoadConnectedRealmTuple,
) (PricelistHistoryDatabase, error) {
	phdBase, ok := phdBases.Databases[tuple.RegionName][tuple.ConnectedRealmId][sotah.UnixTimestamp(
		tuple.LastModified.Unix(),
	)]
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

func (phdBases *PricelistHistoryDatabases) resolveDatabase(
	tuple blizzardv2.LoadConnectedRealmTuple,
) (PricelistHistoryDatabase, error) {
	normalizedTargetDate := sotah.NormalizeTargetDate(tuple.LastModified)
	normalizedTargetTimestamp := sotah.UnixTimestamp(normalizedTargetDate.Unix())

	phdBase, ok := phdBases.Databases[tuple.RegionName][tuple.ConnectedRealmId][normalizedTargetTimestamp]
	if ok {
		return phdBase, nil
	}

	dbPath := pricelistHistoryDatabaseFilePath(
		phdBases.databaseDir,
		tuple.RegionConnectedRealmTuple,
		normalizedTargetTimestamp,
	)
	phdBase, err := newPricelistHistoryDatabase(dbPath, normalizedTargetTimestamp)
	if err != nil {
		return PricelistHistoryDatabase{}, err
	}
	phdBases.Databases[tuple.RegionName][tuple.ConnectedRealmId][normalizedTargetTimestamp] = phdBase

	return phdBase, nil
}

type PricelistHistoryDatabaseShards map[sotah.UnixTimestamp]PricelistHistoryDatabase
