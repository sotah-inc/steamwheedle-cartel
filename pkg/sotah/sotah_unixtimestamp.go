package sotah

import (
	"time"
)

type UnixTimestamps []UnixTimestamp

func (timestamps UnixTimestamps) Before(limit UnixTimestamp) UnixTimestamps {
	out := UnixTimestamps{}
	for _, timestamp := range timestamps {
		if timestamp > limit {
			continue
		}

		out = append(out, timestamp)
	}

	return out
}

func (timestamps UnixTimestamps) IsZero() bool {
	for _, timestamp := range timestamps {
		if timestamp.IsZero() {
			return false
		}
	}

	return true
}

type UnixTimestamp int64

func (timestamp UnixTimestamp) IsZero() bool {
	return timestamp == 0
}

func (timestamp UnixTimestamp) Time() time.Time {
	return time.Unix(int64(timestamp), 0)
}
