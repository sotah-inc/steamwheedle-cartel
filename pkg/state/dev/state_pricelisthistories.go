package dev

import (
	"fmt"
	"github.com/twinj/uuid"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/diskstore"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/metric"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type PricelistHistoriesStateConfig struct {
	MessengerHost string
	MessengerPort int

	DiskStoreCacheDir string

	PricelistHistoriesDatabaseDir string
}

func NewPricelistHistoriesState(config PricelistHistoriesStateConfig) (PricelistHistoriesState, error) {
	phState := PricelistHistoriesState{
		State: state.NewState(uuid.NewV4(), false),
	}
	phState.Statuses = sotah.Statuses{}

	// connecting to the messenger host
	mess, err := messenger.NewMessenger(config.MessengerHost, config.MessengerPort)
	if err != nil {
		return PricelistHistoriesState{}, err
	}
	phState.IO.Messenger = mess

	// initializing a reporter
	phState.IO.Reporter = metric.NewReporter(mess)

	// gathering regions
	regions, err := phState.NewRegions()
	if err != nil {
		return PricelistHistoriesState{}, err
	}
	phState.Regions = regions

	// gathering statuses
	for _, reg := range phState.Regions {
		status, err := phState.IO.Messenger.NewStatus(reg)
		if err != nil {
			return PricelistHistoriesState{}, err
		}

		phState.Statuses[reg.Name] = status
	}

	// ensuring database paths exist
	databasePaths := []string{}
	for regionName, status := range phState.Statuses {
		for _, realm := range status.Realms {
			databasePaths = append(databasePaths, fmt.Sprintf(
				"%s/pricelist-histories/%s/%s",
				config.PricelistHistoriesDatabaseDir,
				regionName,
				realm.Slug,
			))
		}
	}
	if err := util.EnsureDirsExist(databasePaths); err != nil {
		return PricelistHistoriesState{}, err
	}

	// establishing a store
	logging.Info("Connecting to disk store")
	cacheDirs := []string{
		config.DiskStoreCacheDir,
		fmt.Sprintf("%s/pricelist-histories", config.DiskStoreCacheDir),
	}
	for _, reg := range phState.Regions {
		cacheDirs = append(cacheDirs, fmt.Sprintf("%s/pricelist-histories/%s", config.DiskStoreCacheDir, reg.Name))
	}
	if err := util.EnsureDirsExist(cacheDirs); err != nil {
		return PricelistHistoriesState{}, err
	}
	phState.IO.DiskStore = diskstore.NewDiskStore(config.DiskStoreCacheDir)

	return phState, nil
}

type PricelistHistoriesState struct {
	state.State

	Regions  sotah.RegionList
	Statuses sotah.Statuses
}
