package sotah

type StatusTimestamps map[string]UnixTimestamp

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
