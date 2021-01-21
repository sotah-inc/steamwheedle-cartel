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

func NormalizeToDay(targetTimestamp UnixTimestamp) UnixTimestamp {
	targetDate := time.Unix(int64(targetTimestamp), 0)
	nearestDayStartOffset := targetDate.Second() + targetDate.Minute()*60 + targetDate.Hour()*60*60
	return UnixTimestamp(targetDate.Unix() - int64(nearestDayStartOffset))
}

func NormalizeToHour(targetTimestamp UnixTimestamp) UnixTimestamp {
	targetDate := time.Unix(int64(targetTimestamp), 0)
	nearestHourStartOffset := targetDate.Second() + targetDate.Minute()*60
	return UnixTimestamp(targetDate.Unix() - int64(nearestHourStartOffset))
}

type UnixTimestamp int64

func (timestamp UnixTimestamp) IsZero() bool {
	return timestamp == 0
}

func (timestamp UnixTimestamp) Time() time.Time {
	return time.Unix(int64(timestamp), 0)
}
