package fn

import (
	"github.com/twinj/uuid"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/hell"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
)

type GetAreaMapStateConfig struct {
	ProjectId string
}

func NewGetAreaMapState(config GetAreaMapStateConfig) (GetAreaMapState, error) {
	// establishing an initial state
	sta := GetAreaMapState{
		State: state.NewState(uuid.NewV4(), true),
	}

	var err error

	// connecting to hell
	sta.IO.HellClient, err = hell.NewClient(config.ProjectId)
	if err != nil {
		logging.WithField("error", err.Error()).Fatal("failed to connect to firebase")

		return GetAreaMapState{}, err
	}

	return sta, nil
}

type GetAreaMapState struct {
	state.State
}

func (sta GetAreaMapState) Run(id int) (bool, error) {
	areaMap, err := sta.IO.HellClient.GetAreaMap(gameversions.Retail, sotah.AreaMapId(id))
	if err != nil {
		return false, err
	}

	return areaMap != nil, nil
}
