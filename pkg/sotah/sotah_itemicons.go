package sotah

import "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

type IconIdsMap map[IconName]blizzardv2.ItemIds

func (iMap IconIdsMap) Append(name IconName, id blizzardv2.ItemId) IconIdsMap {
	if _, ok := iMap[name]; !ok {
		iMap[name] = blizzardv2.ItemIds{}
	}

	iMap[name] = append(iMap[name], id)

	return iMap
}
