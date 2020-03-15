package hell

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/state"
)

func getAreaMapCollectionName(version gameversions.GameVersion) string {
	return fmt.Sprintf("games/%s/areamaps", version)
}

func getAreaMapDocumentName(version gameversions.GameVersion, id sotah.AreaMapId) string {
	return fmt.Sprintf("%s/%d", getAreaMapCollectionName(version), id)
}

type AreaMap struct {
	State state.State `firestore:"state"`
}

func (c Client) GetAreaMap(gameVersion gameversions.GameVersion, id sotah.AreaMapId) (*AreaMap, error) {
	areaMapRef := c.Doc(getAreaMapDocumentName(gameVersion, id))

	docsnap, err := areaMapRef.Get(c.Context)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}

		return nil, err
	}

	var areaMap AreaMap
	if err := docsnap.DataTo(&areaMap); err != nil {
		return nil, err
	}

	return &areaMap, nil
}
