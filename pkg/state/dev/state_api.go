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

type APIStateConfig struct {
	SotahConfig     sotah.Config
	MessengerConfig struct {
		Hostname string
		Port     int
	}
	DiskStoreCacheDir string
	BlizzardConfig    struct {
		ClientId     string
		ClientSecret string
	}
	DatabaseConfig struct {
		ItemsDir    string
		TokensDir   string
		AreaMapsDir string
	}
}

func NewAPIState(config APIStateConfig) (*APIState, error) {
	// establishing an initial state
	sta := APIState{State: state.NewState(uuid.NewV4(), false)}

	// narrowing regions list
	regions := config.SotahConfig.FilterInRegions(config.SotahConfig.Regions)

	var err error
	sta.diskStore, err = func() (diskstore.DiskStore, error) {
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

		return nil, err
	}

	// connecting to the messenger host
	sta.messenger, err = messenger.NewMessenger(config.MessengerConfig.Hostname, config.MessengerConfig.Port)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to connect to messenger")

		return nil, err
	}

	// initializing a reporter
	sta.reporter = metric.NewReporter(sta.messenger)

	// connecting a new blizzard client
	sta.BlizzardState = state.BlizzardState{}
	sta.BlizzardState.BlizzardClient, err = blizzardv2.NewClient(config.BlizzardConfig.ClientId, config.BlizzardConfig.ClientSecret)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise blizzard-client")

		return nil, err
	}

	// gathering region-state
	sta.RegionState, err = state.NewRegionState(state.NewRegionStateOptions{
		BlizzardState: sta.BlizzardState,
		Regions:       regions,
		Messenger:     sta.messenger,
	})

	// gathering boot-state
	sta.BootState, err = state.NewBootState(state.NewBootStateOptions{
		BlizzardState: sta.BlizzardState,
		Messenger:     sta.messenger,
		Regions:       regions,
		Expansions:    config.SotahConfig.Expansions,
		Professions:   config.SotahConfig.Professions,
		ItemBlacklist: config.SotahConfig.ItemBlacklist,
	})

	// loading the items database
	sta.itemsDatabase, err = database.NewItemsDatabase(config.DatabaseConfig.ItemsDir)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise items-database")

		return nil, err
	}

	// loading the tokens database
	sta.tokensDatabase, err = database.NewTokensDatabase(config.DatabaseConfig.TokensDir)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise tokens-database")

		return nil, err
	}

	// loading the area-maps database
	sta.areaMapsDatabase, err = database.NewAreaMapsDatabase(config.DatabaseConfig.AreaMapsDir)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise area-maps-database")

		return nil, err
	}

	// resolving states
	sta.ItemsState = state.ItemsState{Messenger: sta.messenger, ItemsDatabase: sta.itemsDatabase}
	sta.AreaMapsState = state.AreaMapsState{Messenger: sta.messenger, AreaMapsDatabase: sta.areaMapsDatabase}
	sta.TokensState = state.TokensState{
		BlizzardState:  sta.BlizzardState,
		Messenger:      sta.messenger,
		TokensDatabase: sta.tokensDatabase,
		Reporter:       sta.reporter,
	}
	sta.DiskAuctionsState = state.DiskAuctionsState{
		BlizzardState: sta.BlizzardState,
		RegionsState:  sta.RegionState,
		DiskStore:     sta.diskStore,
	}

	// establishing listeners
	sta.Listeners = state.NewListeners(state.NewSubjectListeners([]state.SubjectListeners{
		sta.ItemsState.GetListeners(),
		sta.AreaMapsState.GetListeners(),
		sta.TokensState.GetListeners(),
		sta.RegionState.GetListeners(),
	}))

	return &sta, nil
}

type APIState struct {
	state.State
	BlizzardState     state.BlizzardState
	ItemsState        state.ItemsState
	AreaMapsState     state.AreaMapsState
	TokensState       state.TokensState
	RegionState       state.RegionsState
	DiskAuctionsState state.DiskAuctionsState
	BootState         state.BootState

	// io
	areaMapsDatabase database.AreaMapsDatabase
	itemsDatabase    database.ItemsDatabase
	tokensDatabase   database.TokensDatabase
	diskStore        diskstore.DiskStore
	messenger        messenger.Messenger
	reporter         metric.Reporter
}
