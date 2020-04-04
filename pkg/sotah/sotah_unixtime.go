package sotah

import (
	"encoding/binary"
	"time"
)

type UnixTime struct {
	time.Time
}

func (t UnixTime) MarshalJSON() ([]byte, error) {
	out := make([]byte, 8)
	binary.LittleEndian.PutUint64(out, uint64(t.Unix()))

	return out, nil
}
