package dev

import (
	"errors"
	"fmt"

	"github.com/twinj/uuid"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/diskstore"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/metric"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type ApiStateDatabaseConfig struct {
	ItemsDir    string
	TokensDir   string
	AreaMapsDir string
}

type ApiStateConfig struct {
	SotahConfig       sotah.Config
	MessengerConfig   messenger.Config
	DiskStoreCacheDir string
	BlizzardConfig    blizzardv2.ClientConfig
	DatabaseConfig    ApiStateDatabaseConfig
}

func NewAPIState(config ApiStateConfig) (ApiState, error) {
	// establishing an initial state
	sta := ApiState{State: state.State{RunID: uuid.NewV4(), Listeners: nil, BusListeners: nil}}

	// narrowing regions list
	regions := config.SotahConfig.FilterInRegions(config.SotahConfig.Regions)

	diskStore, err := func() (diskstore.DiskStore, error) {
		if config.DiskStoreCacheDir == "" {
			logging.WithField("disk-store-cache-dir", config.DiskStoreCacheDir).Error("disk-store-cache-dir was blank")

			return diskstore.DiskStore{}, errors.New("disk-store-cache-dir should not be blank")
		}

		cacheDirs := []string{
			config.DiskStoreCacheDir,
			fmt.Sprintf("%s/items", config.DiskStoreCacheDir),
			fmt.Sprintf("%s/auctions", config.DiskStoreCacheDir),
			fmt.Sprintf("%s/databases", config.DiskStoreCacheDir),
		}
		for _, reg := range regions {
			cacheDirs = append(cacheDirs, fmt.Sprintf("%s/auctions/%s", config.DiskStoreCacheDir, reg.Name))
		}

		if err := util.EnsureDirsExist(cacheDirs); err != nil {
			return diskstore.DiskStore{}, err
		}

		return diskstore.NewDiskStore(config.DiskStoreCacheDir), nil
	}()
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise disk-store")

		return ApiState{}, err
	}

	// connecting to the messenger host
	mess, err := messenger.NewMessenger(config.MessengerConfig)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to connect to messenger")

		return ApiState{}, err
	}

	// connecting a new blizzard client
	sta.BlizzardState, err = state.NewBlizzardState(config.BlizzardConfig)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise blizzard-client")

		return ApiState{}, err
	}

	// gathering region-state
	sta.RegionState, err = state.NewRegionState(state.NewRegionStateOptions{
		BlizzardState: sta.BlizzardState,
		Regions:       regions,
		Messenger:     mess,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to establish region-state")

		return ApiState{}, err
	}

	// gathering boot-state
	sta.BootState, err = state.NewBootState(state.NewBootStateOptions{
		BlizzardState: sta.BlizzardState,
		Messenger:     mess,
		Regions:       regions,
		Expansions:    config.SotahConfig.Expansions,
		Professions:   config.SotahConfig.Professions,
		ItemBlacklist: config.SotahConfig.ItemBlacklist,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to establish boot-state")

		return ApiState{}, err
	}

	// loading the items database
	itemsDatabase, err := database.NewItemsDatabase(config.DatabaseConfig.ItemsDir)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise items-database")

		return ApiState{}, err
	}

	// loading the tokens database
	tokensDatabase, err := database.NewTokensDatabase(config.DatabaseConfig.TokensDir)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise tokens-database")

		return ApiState{}, err
	}

	// loading the area-maps database
	areaMapsDatabase, err := database.NewAreaMapsDatabase(config.DatabaseConfig.AreaMapsDir)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise area-maps-database")

		return ApiState{}, err
	}

	// resolving states
	sta.ItemsState = state.ItemsState{Messenger: mess, ItemsDatabase: itemsDatabase}
	sta.AreaMapsState = state.AreaMapsState{Messenger: mess, AreaMapsDatabase: areaMapsDatabase}
	sta.TokensState = state.TokensState{
		BlizzardState:  sta.BlizzardState,
		Messenger:      mess,
		TokensDatabase: tokensDatabase,
		Reporter:       metric.NewReporter(mess),
	}
	sta.DiskAuctionsState = state.DiskAuctionsState{
		BlizzardState: sta.BlizzardState,
		RegionsState:  sta.RegionState,
		DiskStore:     diskStore,
		ItemsDatabase: itemsDatabase,
	}

	// establishing listeners
	sta.Listeners = state.NewListeners(state.NewSubjectListeners([]state.SubjectListeners{}))

	return sta, nil
}

type ApiState struct {
	state.State
	BlizzardState     state.BlizzardState
	ItemsState        state.ItemsState
	AreaMapsState     state.AreaMapsState
	TokensState       state.TokensState
	RegionState       *state.RegionsState
	DiskAuctionsState state.DiskAuctionsState
	BootState         state.BootState
}
