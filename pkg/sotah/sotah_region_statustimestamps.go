package sotah

import (
	"encoding/json"

	"github.com/sirupsen/logrus"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/statuskinds"
)

type StatusTimestamps map[statuskinds.StatusKind]UnixTimestamp

func (timestamps StatusTimestamps) ToList() UnixTimestamps {
	out := make(UnixTimestamps, len(timestamps))
	i := 0
	for _, timestamp := range timestamps {
		out[i] = timestamp

		i += 1
	}

	return out
}

func (timestamps StatusTimestamps) IsZero() bool {
	logging.WithFields(logrus.Fields{
		"timestamps":      timestamps,
		"timestamps-list": timestamps.ToList(),
		"is-zero":         timestamps.ToList().IsZero(),
	}).Info("checking timestamps")

	return timestamps.ToList().IsZero()
}

func (timestamps StatusTimestamps) Merge(
	in StatusTimestamps,
) StatusTimestamps {
	for k, timestamp := range in {
		if timestamp.IsZero() {
			continue
		}

		timestamps[k] = timestamp
	}

	return timestamps
}

func (timestamps StatusTimestamps) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(timestamps)
}
