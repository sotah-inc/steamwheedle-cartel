package blizzardv2

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewVersionItemsMap(data []byte) (VersionItemsMap, error) {
	gzipDecoded, err := util.GzipDecode(data)
	if err != nil {
		return VersionItemsMap{}, err
	}

	out := VersionItemsMap{}
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return VersionItemsMap{}, err
	}

	return out, nil
}

type VersionItemsMap map[gameversion.GameVersion]ItemIds

func (viMap VersionItemsMap) Insert(
	gameVersion gameversion.GameVersion,
	ids ItemIds,
) VersionItemsMap {
	viMap[gameVersion] = viMap.Resolve(gameVersion).Merge(ids)

	return viMap
}

func (viMap VersionItemsMap) Resolve(gameVersion gameversion.GameVersion) ItemIds {
	out, ok := viMap[gameVersion]
	if !ok {
		return ItemIds{}
	}

	return out
}

func (viMap VersionItemsMap) IsZero() bool {
	for _, ids := range viMap {
		if len(ids) > 0 {
			return false
		}
	}

	return true
}

func (viMap VersionItemsMap) EncodeForDelivery() ([]byte, error) {
	jsonEncoded, err := json.Marshal(viMap)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncoded)
}
