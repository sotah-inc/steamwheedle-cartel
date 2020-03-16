package fn

import (
	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"github.com/twinj/uuid"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/hell"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/store"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/store/regions"
)

type LoadAreaMapsStateConfig struct {
	ProjectId string
}

func NewLoadAreaMapsState(config LoadAreaMapsStateConfig) (LoadAreaMapsState, error) {
	// establishing an initial state
	sta := LoadAreaMapsState{
		State: state.NewState(uuid.NewV4(), true),
	}

	var err error

	// connecting to hell
	sta.IO.HellClient, err = hell.NewClient(config.ProjectId)
	if err != nil {
		logging.Fatalf("Failed to connect to firebase: %s", err.Error())

		return LoadAreaMapsState{}, err
	}

	// initializing a store client
	sta.IO.StoreClient, err = store.NewClient(config.ProjectId)
	if err != nil {
		logging.Fatalf("Failed to create new store client: %s", err.Error())

		return LoadAreaMapsState{}, err
	}

	sta.bootBase = store.NewBootBase(sta.IO.StoreClient, regions.USCentral1)
	sta.bootBucket, err = sta.bootBase.GetFirmBucket()
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Fatal("failed to get firm bucket for boot-base")

		return LoadAreaMapsState{}, err
	}

	sta.areaMapsBase = store.NewAreaMapsBase(sta.IO.StoreClient, regions.USCentral1, gameversions.Retail)
	sta.areaMapsBucket, err = sta.areaMapsBase.GetFirmBucket()
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Fatal("failed to get firm bucket for area-maps-base")

		return LoadAreaMapsState{}, err
	}

	return sta, nil
}

type LoadAreaMapsState struct {
	state.State

	bootBase   store.BootBase
	bootBucket *storage.BucketHandle

	areaMapsBase   store.AreaMapsBase
	areaMapsBucket *storage.BucketHandle
}

func (sta LoadAreaMapsState) Run() error {
	return sta.Reset()
}
