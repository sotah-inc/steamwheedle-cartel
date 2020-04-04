package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func NewRegionRealmTupleFromRealm(r Realm) RegionRealmTuple {
	return RegionRealmTuple{
		RegionName: r.Region.Name,
		RealmSlug:  r.Slug,
	}
}

func NewRegionRealmTuple(data []byte) (RegionRealmTuple, error) {
	var out RegionRealmTuple
	if err := json.Unmarshal(data, &out); err != nil {
		return RegionRealmTuple{}, err
	}

	return out, nil
}

type RegionRealmTuple struct {
	RegionName blizzardv2.RegionName `json:"region_name"`
	RealmSlug  blizzardv2.RealmSlug  `json:"realm_slug"`
}

func (tuple RegionRealmTuple) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(tuple)
	if err != nil {
		return "", err
	}

	return string(jsonEncoded), nil
}

func NewRegionRealmTuples(data string) (RegionRealmTuples, error) {
	var out RegionRealmTuples
	if err := json.Unmarshal([]byte(data), &out); err != nil {
		return RegionRealmTuples{}, err
	}

	return out, nil
}

type RegionRealmTuples []RegionRealmTuple

func (tuples RegionRealmTuples) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(tuples)
	if err != nil {
		return "", err
	}

	return string(jsonEncoded), nil
}

func (tuples RegionRealmTuples) ToRegionRealmSlugs() RegionRealmSlugs {
	out := RegionRealmSlugs{}
	for _, tuple := range tuples {
		next := func() []blizzardv2.RealmSlug {
			result, ok := out[tuple.RegionName]
			if ok {
				return result
			}

			return []blizzardv2.RealmSlug{}
		}()

		next = append(next, tuple.RealmSlug)
		out[tuple.RegionName] = next
	}

	return out
}

func NewRegionRealmTimestampTuples(data string) (RegionRealmTimestampTuples, error) {
	var out RegionRealmTimestampTuples
	if err := json.Unmarshal([]byte(data), &out); err != nil {
		return RegionRealmTimestampTuples{}, err
	}

	return out, nil
}

type RegionRealmTimestampTuples []RegionRealmTimestampTuple

func (tuples RegionRealmTimestampTuples) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(tuples)
	if err != nil {
		return "", err
	}

	return string(jsonEncoded), nil
}

func (tuples RegionRealmTimestampTuples) ToRegionRealmSlugs() RegionRealmSlugs {
	out := RegionRealmSlugs{}
	for _, tuple := range tuples {
		next := func() []blizzardv2.RealmSlug {
			result, ok := out[tuple.RegionName]
			if ok {
				return result
			}

			return []blizzardv2.RealmSlug{}
		}()

		next = append(next, tuple.RealmSlug)
		out[tuple.RegionName] = next
	}

	return out
}
