package database

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type PricelistHistoryDatabaseShards map[sotah.UnixTimestamp]PricelistHistoryDatabase

func (shards PricelistHistoryDatabaseShards) Before(
	limit sotah.UnixTimestamp,
	inclusive bool,
) PricelistHistoryDatabaseShards {
	out := PricelistHistoryDatabaseShards{}
	for timestamp, phdBase := range shards {
		if timestamp < limit || timestamp == limit && inclusive {
			out[timestamp] = phdBase
		}
	}

	logging.WithFields(logrus.Fields{
		"out":   len(out),
		"limit": limit,
	}).Info("found before")

	return out
}

func (shards PricelistHistoryDatabaseShards) After(
	limit sotah.UnixTimestamp,
	inclusive bool,
) PricelistHistoryDatabaseShards {
	out := PricelistHistoryDatabaseShards{}
	for timestamp, phdBase := range shards {
		if inclusive && timestamp == limit || timestamp > limit {
			out[timestamp] = phdBase
		}
	}

	logging.WithFields(logrus.Fields{
		"out":   len(out),
		"limit": limit,
	}).Info("found after")

	return out
}

func (shards PricelistHistoryDatabaseShards) Between(
	lowerLimit sotah.UnixTimestamp,
	upperLimit sotah.UnixTimestamp,
) PricelistHistoryDatabaseShards {
	return shards.After(lowerLimit, true).Before(upperLimit, true)
}
