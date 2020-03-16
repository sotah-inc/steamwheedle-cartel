package hell

import (
	"strconv"

	"google.golang.org/api/iterator"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/state"
)

func (c Client) FilterInByState(version gameversions.GameVersion, sotahState state.State) (sotah.AreaMapMap, error) {
	colRef := c.Collection(getAreaMapCollectionName(version))

	out := sotah.AreaMapMap{}
	docIter := colRef.Where("state", "==", sotahState).Documents(c.Context)
	defer docIter.Stop()
	for {
		docSnap, err := docIter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}

			return sotah.AreaMapMap{}, err
		}

		var areaMapData AreaMap
		if err := docSnap.DataTo(&areaMapData); err != nil {
			return sotah.AreaMapMap{}, err
		}

		parsedAreaMapId, err := strconv.Atoi(docSnap.Ref.ID)
		if err != nil {
			return sotah.AreaMapMap{}, err
		}

		out[sotah.AreaMapId(parsedAreaMapId)] = sotah.AreaMap{
			Id:             sotah.AreaMapId(parsedAreaMapId),
			State:          areaMapData.State,
			Name:           "",
			NormalizedName: "",
		}
	}

	return sotah.AreaMapMap{}, nil
}
