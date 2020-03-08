package fn

import (
	"cloud.google.com/go/storage"
	"github.com/twinj/uuid"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/hell"
	HellState "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/hell/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/store"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/store/regions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/wowhead"
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
		logging.Fatalf("Failed to get firm bucket: %s", err.Error())

		return LoadAreaMapsState{}, err
	}

	sta.areaMapsBase = store.NewAreaMapsBase(sta.IO.StoreClient, regions.USCentral1, gameversions.Classic)
	sta.areaMapsBucket, err = sta.areaMapsBase.GetFirmBucket()
	if err != nil {
		logging.Fatalf("Failed to get firm bucket: %s", err.Error())

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

func (sta LoadAreaMapsState) Run(filteredZoneIds []int) error {
	// initializing loaders
	baseLoadAreaMapsIn := make(chan store.LoadAreaMapsInJob)
	baseLoadAreaMapsOut := sta.areaMapsBase.LoadAreaMaps(baseLoadAreaMapsIn, sta.areaMapsBucket)
	hellLoadAreaMapsIn := make(chan hell.LoadAreaMapsInJob)
	hellLoadAreaMapsOut := sta.IO.HellClient.LoadAreaMaps(gameversions.Retail, hellLoadAreaMapsIn)

	// spinning up the download worker
	go func() {
		downloadOut := wowhead.DownloadAreaMaps(filteredZoneIds)
		for downloadOutJob := range downloadOut {
			if downloadOutJob.Err != nil {
				logging.WithFields(downloadOutJob.ToLogrusFields()).Error("failed to get firm bucket")

				hellLoadAreaMapsIn <- hell.LoadAreaMapsInJob{
					Id:    downloadOutJob.Id,
					State: HellState.Erroneous,
				}

				continue
			}

			baseLoadAreaMapsIn <- store.LoadAreaMapsInJob{
				AreaId: downloadOutJob.Id,
				Data:   downloadOutJob.Data,
			}
		}

		close(baseLoadAreaMapsIn)
	}()

	// spinning up the base-load worker
	go func() {
		for baseLoadAreaMapsOutJob := range baseLoadAreaMapsOut {
			if baseLoadAreaMapsOutJob.Err != nil {
				logging.WithFields(baseLoadAreaMapsOutJob.ToLogrusFields()).Error("failed to load area-map into base")

				hellLoadAreaMapsIn <- hell.LoadAreaMapsInJob{
					Id:    baseLoadAreaMapsOutJob.AreaId,
					State: HellState.Erroneous,
				}

				continue
			}

			hellLoadAreaMapsIn <- hell.LoadAreaMapsInJob{
				Id:    baseLoadAreaMapsOutJob.AreaId,
				State: HellState.Complete,
			}
		}

		close(hellLoadAreaMapsIn)
	}()

	// waiting for it to drain out
	for hellLoadAreaMapsOutJob := range hellLoadAreaMapsOut {
		if hellLoadAreaMapsOutJob.Err != nil {
			logging.WithFields(hellLoadAreaMapsOutJob.ToLogrusFields()).Error("failed to load area-map into hell")

			continue
		}
	}

	return nil
}
