package fn

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	sotahState "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/state"
)

func (sta LoadAreaMapsState) getCompleteFullAreaMaps() (sotah.AreaMapMap, error) {
	// gathering from store
	parentAreaMaps, err := sta.bootBase.GetParentAreaMaps()
	if err != nil {
		return sotah.AreaMapMap{}, err
	}

	// gathering complete from hell
	completeAreaMaps, err := sta.IO.HellClient.FilterInByState(gameversions.Retail, sotahState.Complete)
	if err != nil {
		return sotah.AreaMapMap{}, err
	}

	// correlating
	out := sotah.AreaMapMap{}
	for id := range completeAreaMaps {
		out[id] = parentAreaMaps[id]
	}

	return out, nil
}

func (sta LoadAreaMapsState) Store() error {
	fullCompletedAreaMaps, err := sta.getCompleteFullAreaMaps()
	if err != nil {
		return err
	}

	if err := sta.areaMapsDb.PersistAreaMaps(fullCompletedAreaMaps); err != nil {
		return err
	}

	found, err := sta.areaMapsDb.AreaMapsQuery(database.AreaMapsQueryRequest{Query: "the"})
	if err != nil {
		return err
	}

	logging.WithField("found", found).Info("Found results")

	return nil
}
