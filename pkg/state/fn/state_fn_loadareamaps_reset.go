package fn

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	sotahState "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/store"
)

func (sta LoadAreaMapsState) Reset() error {
	areaMaps, err := sta.IO.HellClient.FilterInByState(gameversions.Retail, sotahState.Complete)
	if err != nil {
		return err
	}

	// establishing channels
	resetAreaMapsIn := make(chan store.ResetAreaMapsInJob)
	resetAreaMapsOut := sta.areaMapsBase.ResetAreaMaps(resetAreaMapsIn, sta.areaMapsBucket)

	// queueing it up
	go func() {
		for id := range areaMaps {
			resetAreaMapsIn <- store.ResetAreaMapsInJob{
				AreaId: id,
			}
		}

		close(resetAreaMapsIn)
	}()

	// waiting for it to drain out
	for outJob := range resetAreaMapsOut {
		if outJob.Err != nil {
			logging.WithFields(outJob.ToLogrusFields()).Error("failed to reset area-map")

			return outJob.Err
		}
	}

	return nil
}
