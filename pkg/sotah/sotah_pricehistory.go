package sotah

import (
	"encoding/base64"
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

// price-history
func NewPriceHistoryFromBytes(data []byte) (PriceHistory, error) {
	gzipDecoded, err := util.GzipDecode(data)
	if err != nil {
		return PriceHistory{}, err
	}

	out := PriceHistory{}
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return PriceHistory{}, err
	}

	return out, nil
}

type PriceHistory map[UnixTimestamp]Prices

func (pHistory PriceHistory) EncodeForPersistence() (string, error) {
	jsonEncoded, err := json.Marshal(pHistory)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}

func (pHistory PriceHistory) Merge(in PriceHistory) PriceHistory {
	for timestamp, prices := range in {
		pHistory[timestamp] = prices
	}

	return pHistory
}

func (pHistory PriceHistory) Before(
	limit UnixTimestamp,
	inclusive bool,
) PriceHistory {
	out := PriceHistory{}
	for timestamp, prices := range pHistory {
		if timestamp < limit || timestamp == limit && inclusive {
			out[timestamp] = prices
		}
	}

	return out
}

func (pHistory PriceHistory) After(
	limit UnixTimestamp,
	inclusive bool,
) PriceHistory {
	out := PriceHistory{}
	for timestamp, prices := range pHistory {
		if timestamp == limit && inclusive || timestamp > limit {
			out[timestamp] = prices
		}
	}

	return out
}

func (pHistory PriceHistory) Between(
	lowerLimit UnixTimestamp,
	upperLimit UnixTimestamp,
) PriceHistory {
	return pHistory.After(lowerLimit, true).Before(upperLimit, true)
}
