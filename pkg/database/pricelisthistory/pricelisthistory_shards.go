package pricelisthistory

import (
	"errors"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type DatabaseShards map[sotah.UnixTimestamp]Database

func (shards DatabaseShards) Latest() (Database, error) {
	lastTimestamp := sotah.UnixTimestamp(0)
	for timestamp := range shards {
		if lastTimestamp == 0 || timestamp > lastTimestamp {
			lastTimestamp = timestamp
		}
	}

	if lastTimestamp == 0 {
		return Database{}, errors.New("failed to resolve latest database")
	}

	return shards[lastTimestamp], nil
}

func (shards DatabaseShards) Before(
	limit sotah.UnixTimestamp,
	inclusive bool,
) DatabaseShards {
	out := DatabaseShards{}
	for timestamp, phdBase := range shards {
		if timestamp < limit || timestamp == limit && inclusive {
			out[timestamp] = phdBase
		}
	}

	return out
}

func (shards DatabaseShards) After(
	limit sotah.UnixTimestamp,
	inclusive bool,
) DatabaseShards {
	out := DatabaseShards{}
	for timestamp, phdBase := range shards {
		if inclusive && timestamp == limit || timestamp > limit {
			out[timestamp] = phdBase
		}
	}

	return out
}

func (shards DatabaseShards) Between(
	lowerLimit sotah.UnixTimestamp,
	upperLimit sotah.UnixTimestamp,
) DatabaseShards {
	return shards.After(lowerLimit, true).Before(upperLimit, true)
}
