package bus

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
)

func NewRegionRealmTimestampTuples(data string) (RegionRealmTimestampTuples, error) {
	base64Decoded, err := base64.RawStdEncoding.DecodeString(data)
	if err != nil {
		return RegionRealmTimestampTuples{}, err
	}

	gzipDecoded, err := util.GzipDecode(base64Decoded)
	if err != nil {
		return RegionRealmTimestampTuples{}, err
	}

	var out RegionRealmTimestampTuples
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return RegionRealmTimestampTuples{}, err
	}

	return out, nil
}

type RegionRealmTimestampTuples []RegionRealmTimestampTuple

func (s RegionRealmTimestampTuples) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(gzipEncoded), nil
}

func (s RegionRealmTimestampTuples) ToMessages() ([]Message, error) {
	out := []Message{}
	for _, tuple := range s {
		msg := NewMessage()
		msg.ReplyToId = fmt.Sprintf("%s-%s", tuple.RegionName, tuple.RealmSlug)

		job := LoadRegionRealmTimestampsInJob{
			RegionName:      tuple.RegionName,
			RealmSlug:       tuple.RealmSlug,
			TargetTimestamp: tuple.TargetTimestamp,
		}
		data, err := job.EncodeForDelivery()
		if err != nil {
			return []Message{}, err
		}
		msg.Data = data

		out = append(out, msg)
	}

	return out, nil
}

func (s RegionRealmTimestampTuples) ToRegionRealmSlugs() map[blizzard.RegionName][]blizzard.RealmSlug {
	out := map[blizzard.RegionName][]blizzard.RealmSlug{}
	for _, tuple := range s {
		next := func() []blizzard.RealmSlug {
			result, ok := out[blizzard.RegionName(tuple.RegionName)]
			if ok {
				return result
			}

			return []blizzard.RealmSlug{}
		}()

		next = append(next, blizzard.RealmSlug(tuple.RealmSlug))
		out[blizzard.RegionName(tuple.RegionName)] = next
	}

	return out
}

func NewRegionRealmTimestampTuple(data string) (RegionRealmTimestampTuple, error) {
	base64Decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return RegionRealmTimestampTuple{}, err
	}

	gzipDecoded, err := util.GzipDecode(base64Decoded)
	if err != nil {
		return RegionRealmTimestampTuple{}, err
	}

	var out RegionRealmTimestampTuple
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return RegionRealmTimestampTuple{}, err
	}

	return out, nil
}

type RegionRealmTimestampTuple struct {
	RegionName                string   `json:"region_name"`
	RealmSlug                 string   `json:"realm_slug"`
	TargetTimestamp           int      `json:"target_timestamp"`
	NormalizedTargetTimestamp int      `json:"normalized_target_timestamp"`
	ItemIds                   []int    `json:"item_ids"`
	OwnerNames                []string `json:"owner_names"`
}

func (t RegionRealmTimestampTuple) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(t)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}

func (t RegionRealmTimestampTuple) Bare() RegionRealmTimestampTuple {
	return RegionRealmTimestampTuple{
		RegionName:      t.RegionName,
		RealmSlug:       t.RealmSlug,
		TargetTimestamp: t.TargetTimestamp,
	}
}
