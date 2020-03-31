package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
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

func NewSkeletonRealm(regionName blizzardv2.RegionName, realmSlug blizzardv2.RealmSlug) Realm {
	return Realm{
		Region: Region{Name: regionName},
		Realm:  blizzard.Realm{Slug: realmSlug},
	}
}

type Realm struct {
	blizzard.Realm
	Region Region `json:"region"`
}

type RegionRealms map[blizzardv2.RegionName]Realms

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
		out[regionName] = make([]blizzardv2.RealmSlug, len(realms))
		i := 0
		for _, realm := range realms {
			out[regionName][i] = realm.Slug

			i++
		}
	}

	return out
}

type RegionRealmSlugs map[blizzardv2.RegionName][]blizzardv2.RealmSlug

type RegionRealmMap map[blizzardv2.RegionName]RealmMap

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
		out[regionName] = make([]blizzardv2.RealmSlug, len(realmsMap))
		i := 0
		for realmSlug := range realmsMap {
			out[regionName][i] = realmSlug

			i++
		}
	}

	return out
}

type RealmMap map[blizzardv2.RealmSlug]Realm

func (rMap RealmMap) ToRealms() Realms {
	out := Realms{}
	for _, realm := range rMap {
		out = append(out, realm)
	}

	return out
}
