package sotah

import (
	"encoding/json"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
)

func NewRealms(reg Region, blizzRealms []blizzard.Realm) Realms {
	reas := make([]Realm, len(blizzRealms))
	for i, rea := range blizzRealms {
		reas[i] = Realm{rea, reg}
	}

	return reas
}

type Realms []Realm

func (realms Realms) ToRealmMap() RealmMap {
	out := RealmMap{}
	for _, realm := range realms {
		out[realm.Slug] = realm
	}

	return out
}

type RegionRealmModificationDates map[blizzard.RegionName]map[blizzard.RealmSlug]RealmModificationDates

func (d RegionRealmModificationDates) Get(
	regionName blizzard.RegionName,
	realmSlug blizzard.RealmSlug,
) RealmModificationDates {
	realmsModDates, ok := d[regionName]
	if !ok {
		return RealmModificationDates{}
	}

	realmModDates, ok := realmsModDates[realmSlug]
	if !ok {
		return RealmModificationDates{}
	}

	return realmModDates
}

func (d RegionRealmModificationDates) Set(
	regionName blizzard.RegionName,
	realmSlug blizzard.RealmSlug,
	modDates RealmModificationDates,
) RegionRealmModificationDates {
	realmsModDates, ok := d[regionName]
	if !ok {
		d[regionName] = map[blizzard.RealmSlug]RealmModificationDates{realmSlug: modDates}

		return d
	}

	realmsModDates[realmSlug] = modDates
	d[regionName] = realmsModDates

	return d
}

func (d RegionRealmModificationDates) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(d)
}

type RealmModificationDates struct {
	Downloaded                 int64 `json:"downloaded"`
	LiveAuctionsReceived       int64 `json:"live_auctions_received"`
	PricelistHistoriesReceived int64 `json:"pricelist_histories_received"`
}

func NewSkeletonRealm(regionName blizzard.RegionName, realmSlug blizzard.RealmSlug) Realm {
	return Realm{
		Region: Region{Name: regionName},
		Realm:  blizzard.Realm{Slug: realmSlug},
	}
}

type Realm struct {
	blizzard.Realm
	Region Region `json:"region"`
}

func (r Realm) EncodeForStorage() ([]byte, error) {
	jsonEncoded, err := json.Marshal(r)
	if err != nil {
		return []byte{}, err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return []byte{}, err
	}

	return gzipEncoded, nil
}

type RegionRealms map[blizzard.RegionName]Realms

func (regionRealms RegionRealms) TotalRealms() int {
	out := 0
	for _, realms := range regionRealms {
		out += len(realms)
	}

	return out
}

func (regionRealms RegionRealms) ToRegionRealmSlugs() RegionRealmSlugs {
	out := RegionRealmSlugs{}

	for regionName, realms := range regionRealms {
		out[regionName] = make([]blizzard.RealmSlug, len(realms))
		i := 0
		for _, realm := range realms {
			out[regionName][i] = realm.Slug

			i++
		}
	}

	return out
}

type RegionRealmSlugs map[blizzard.RegionName][]blizzard.RealmSlug

type RegionRealmMap map[blizzard.RegionName]RealmMap

func (regionRealmMap RegionRealmMap) ToRegionRealms() RegionRealms {
	out := RegionRealms{}
	for regionName, realmMap := range regionRealmMap {
		out[regionName] = realmMap.ToRealms()
	}

	return out
}

func (regionRealmMap RegionRealmMap) ToRegionRealmSlugs() RegionRealmSlugs {
	out := RegionRealmSlugs{}

	for regionName, realmsMap := range regionRealmMap {
		out[regionName] = make([]blizzard.RealmSlug, len(realmsMap))
		i := 0
		for realmSlug := range realmsMap {
			out[regionName][i] = realmSlug

			i++
		}
	}

	return out
}

type RealmMap map[blizzard.RealmSlug]Realm

func (rMap RealmMap) ToRealms() Realms {
	out := Realms{}
	for _, realm := range rMap {
		out = append(out, realm)
	}

	return out
}

type RealmTimestampMap map[blizzard.RealmSlug]int64

type RegionRealmTimestampMaps map[blizzard.RegionName]RealmTimestampMap

type RealmTimestamps map[blizzard.RealmSlug][]UnixTimestamp

type RegionRealmTimestamps map[blizzard.RegionName]RealmTimestamps

func NewRegionRealmTupleFromRealm(r Realm) RegionRealmTuple {
	return RegionRealmTuple{
		RegionName: string(r.Region.Name),
		RealmSlug:  string(r.Slug),
	}
}

func NewRegionRealmTuple(data string) (RegionRealmTuple, error) {
	var out RegionRealmTuple
	if err := json.Unmarshal([]byte(data), &out); err != nil {
		return RegionRealmTuple{}, err
	}

	return out, nil
}

type RegionRealmTuple struct {
	RegionName string `json:"region_name"`
	RealmSlug  string `json:"realm_slug"`
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

type RegionRealmSummaryTuples []RegionRealmSummaryTuple

func (tuples RegionRealmSummaryTuples) ItemIds() blizzard.ItemIds {
	itemIdsMap := ItemIdsMap{}
	for _, tuple := range tuples {
		for _, id := range tuple.ItemIds {
			itemIdsMap[blizzard.ItemID(id)] = struct{}{}
		}
	}
	out := blizzard.ItemIds{}
	for id := range itemIdsMap {
		out = append(out, id)
	}

	return out
}

func (tuples RegionRealmSummaryTuples) RegionRealmTuples() RegionRealmTuples {
	out := make(RegionRealmTuples, len(tuples))
	for i, tuple := range tuples {
		out[i] = tuple.RegionRealmTuple
	}

	return out
}

func NewRegionRealmSummaryTuple(data string) (RegionRealmSummaryTuple, error) {
	var out RegionRealmSummaryTuple
	if err := json.Unmarshal([]byte(data), &out); err != nil {
		return RegionRealmSummaryTuple{}, err
	}

	return out, nil
}

type RegionRealmSummaryTuple struct {
	RegionRealmTimestampTuple
	ItemIds    []int    `json:"item_ids"`
	OwnerNames []string `json:"owner_names"`
}

func (tuple RegionRealmSummaryTuple) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(tuple)
	if err != nil {
		return "", err
	}

	return string(jsonEncoded), nil
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
