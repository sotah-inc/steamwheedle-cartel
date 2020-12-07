package sotah

import (
	"encoding/base64"
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

// recipe price-history
type RecipePriceHistory map[UnixTimestamp]RecipePrices

func (rpHistory RecipePriceHistory) EncodeForPersistence() (string, error) {
	jsonEncoded, err := json.Marshal(rpHistory)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}

func (rpHistory RecipePriceHistory) Merge(in RecipePriceHistory) RecipePriceHistory {
	for timestamp, prices := range in {
		rpHistory[timestamp] = prices
	}

	return rpHistory
}

func (rpHistory RecipePriceHistory) Before(
	limit UnixTimestamp,
	inclusive bool,
) RecipePriceHistory {
	out := RecipePriceHistory{}
	for timestamp, prices := range rpHistory {
		if timestamp < limit || timestamp == limit && inclusive {
			out[timestamp] = prices
		}
	}

	return out
}

func (rpHistory RecipePriceHistory) After(
	limit UnixTimestamp,
	inclusive bool,
) RecipePriceHistory {
	out := RecipePriceHistory{}
	for timestamp, prices := range rpHistory {
		if timestamp == limit && inclusive || timestamp > limit {
			out[timestamp] = prices
		}
	}

	return out
}

func (rpHistory RecipePriceHistory) Between(
	lowerLimit UnixTimestamp,
	upperLimit UnixTimestamp,
) RecipePriceHistory {
	return rpHistory.After(lowerLimit, true).Before(upperLimit, true)
}
