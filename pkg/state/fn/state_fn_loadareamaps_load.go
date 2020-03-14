package fn

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/hell"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	HellState "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/store"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/wowhead"
)

func (sta LoadAreaMapsState) Load() error {
	// gathering parent-zone-ids
	parentZoneIds, err := sta.bootBase.GetParentZoneIds()
	if err != nil {
		return err
	}

	logging.WithField("parent-zone-ids", len(parentZoneIds)).Info("Found parent-zone-ids")

	filteredZoneIds, err := sta.IO.HellClient.FilterInNonExist(gameversions.Retail, parentZoneIds)
	if err != nil {
		return err
	}

	logging.WithField("filtered-zone-ids", len(filteredZoneIds)).Info("Found filtered-zone-ids")

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
				logging.WithFields(downloadOutJob.ToLogrusFields()).Error("failed to download area-map from wowhead")

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
