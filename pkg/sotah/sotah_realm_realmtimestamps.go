package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type RealmTimestampMap map[blizzardv2.RealmSlug]int64

type RegionRealmTimestampMaps map[blizzardv2.RegionName]RealmTimestampMap

type RealmTimestamps map[blizzardv2.RealmSlug][]UnixTimestamp

type RegionRealmTimestamps map[blizzardv2.RegionName]RealmTimestamps

func NewRegionRealmTimestampTuple(data string) (RegionRealmTimestampTuple, error) {
	var out RegionRealmTimestampTuple
	if err := json.Unmarshal([]byte(data), &out); err != nil {
		return RegionRealmTimestampTuple{}, err
	}

	return out, nil
}

type RegionRealmTimestampTuple struct {
	RegionRealmTuple
	TargetTimestamp int `json:"target_timestamp"`
}

func (tuple RegionRealmTimestampTuple) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(tuple)
	if err != nil {
		return "", err
	}

	return string(jsonEncoded), nil
}

func NewRegionRealmTimestampSizeTuple(data string) (RegionRealmTimestampSizeTuple, error) {
	var out RegionRealmTimestampSizeTuple
	if err := json.Unmarshal([]byte(data), &out); err != nil {
		return RegionRealmTimestampSizeTuple{}, err
	}

	return out, nil
}

type RegionRealmTimestampSizeTuple struct {
	RegionRealmTimestampTuple
	SizeBytes int `json:"size_bytes"`
}

func (tuple RegionRealmTimestampSizeTuple) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(tuple)
	if err != nil {
		return "", err
	}

	return string(jsonEncoded), nil
}
