package database

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type PricelistHistoryDatabases map[blizzardv2.RegionName]map[blizzardv2.ConnectedRealmId]PricelistHistoryDatabaseShards

type PricelistHistoryDatabaseShards map[sotah.UnixTimestamp]PricelistHistoryDatabase
