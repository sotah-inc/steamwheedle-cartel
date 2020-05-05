package dev

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake"

	"github.com/twinj/uuid"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseCollector "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/collector/base"
	DiskCollector "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/collector/disk"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type ApiStateDatabaseConfig struct {
	ItemsDir            string
	TokensDir           string
	AreaMapsDir         string
	LiveAuctionsDir     string
	PricelistHistoryDir string
}

type ApiStateConfig struct {
	SotahConfig       sotah.Config
	MessengerConfig   messenger.Config
	DiskStoreCacheDir string
	BlizzardConfig    blizzardv2.ClientConfig
	DatabaseConfig    ApiStateDatabaseConfig
	UseGCloud         bool
}

func (c ApiStateConfig) ToDirList() []string {
	out := []string{
		c.DatabaseConfig.AreaMapsDir,
		c.DatabaseConfig.ItemsDir,
		c.DatabaseConfig.TokensDir,
		c.DatabaseConfig.LiveAuctionsDir,
		fmt.Sprintf("%s/live-auctions", c.DatabaseConfig.LiveAuctionsDir),
	}

	for _, reg := range c.SotahConfig.FilterInRegions(c.SotahConfig.Regions) {
		out = append(out, fmt.Sprintf("%s/live-auctions/%s", c.DatabaseConfig.LiveAuctionsDir, reg.Name))
	}

	return out
}

func NewAPIState(config ApiStateConfig) (ApiState, error) {
	// establishing an initial state
	sta := ApiState{State: state.State{RunID: uuid.NewV4(), Listeners: nil, BusListeners: nil}}

	// narrowing regions list
	regions := config.SotahConfig.FilterInRegions(config.SotahConfig.Regions)

	// resolving primary-region
	primaryRegion, err := regions.GetPrimaryRegion()
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to resolve primary-region")

		return ApiState{}, err
	}

	// ensuring related dirs exist
	if err := util.EnsureDirsExist(config.ToDirList()); err != nil {
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

	// deriving lake-client
	lakeClient, err := lake.NewClient(lake.NewClientOptions{
		UseGCloud: config.UseGCloud,
		CacheDir:  config.DiskStoreCacheDir,
		ResolveItems: func(ids blizzardv2.ItemIds) chan blizzardv2.GetItemsOutJob {
			return sta.BlizzardState.ResolveItems(primaryRegion, ids)
		},
		ResolveItemMedias: sta.BlizzardState.ResolveItemMedias,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise lake-client")

		return ApiState{}, err
	}

	// gathering region state
	sta.RegionState, err = state.NewRegionState(state.NewRegionStateOptions{
		BlizzardState:            sta.BlizzardState,
		Regions:                  regions,
		Messenger:                mess,
		RegionRealmSlugWhitelist: config.SotahConfig.Whitelist,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to establish region state")

		return ApiState{}, err
	}

	// gathering boot state
	sta.BootState, err = state.NewBootState(state.NewBootStateOptions{
		BlizzardState: sta.BlizzardState,
		Messenger:     mess,
		Regions:       regions,
		Expansions:    config.SotahConfig.Expansions,
		Professions:   config.SotahConfig.Professions,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to establish boot state")

		return ApiState{}, err
	}

	sta.ItemsState, err = state.NewItemsState(state.NewItemsStateOptions{
		LakeClient:       lakeClient,
		Messenger:        mess,
		ItemsDatabaseDir: config.DatabaseConfig.ItemsDir,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise items state")

		return ApiState{}, err
	}

	// loading the tokens state
	sta.TokensState, err = state.NewTokensState(state.NewTokensStateOptions{
		BlizzardState:     sta.BlizzardState,
		Messenger:         mess,
		TokensDatabaseDir: config.DatabaseConfig.TokensDir,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise tokens state")

		return ApiState{}, err
	}

	// loading the area-maps state
	sta.AreaMapsState, err = state.NewAreaMapsState(mess, config.DatabaseConfig.AreaMapsDir)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise area-maps state")

		return ApiState{}, err
	}

	// resolving DiskCollector-auctions state
	sta.Collector = DiskCollector.NewClient(DiskCollector.ClientOptions{
		ResolveAuctions: func() chan blizzardv2.GetAuctionsJob {
			return sta.BlizzardState.ResolveAuctions(sta.RegionState.RegionComposites.ToDownloadTuples())
		},
		ReceiveRegionTimestamps: sta.RegionState.ReceiveTimestamps,
		LakeClient:              lakeClient,
		MessengerClient:         mess,
	})

	// resolving live-auctions state
	sta.LiveAuctionsState, err = state.NewLiveAuctionsState(state.NewLiveAuctionsStateOptions{
		Messenger:                mess,
		LakeClient:               lakeClient,
		LiveAuctionsDatabasesDir: config.DatabaseConfig.LiveAuctionsDir,
		Tuples:                   sta.RegionState.RegionComposites.ToTuples(),
		ReceiveRegionTimestamps:  sta.RegionState.ReceiveTimestamps,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise live-auctions state")

		return ApiState{}, err
	}

	// resolving pricelist-history state
	sta.PricelistHistoryState, err = state.NewPricelistHistoryState(state.NewPricelistHistoryStateOptions{
		Messenger:                    mess,
		LakeClient:                   lakeClient,
		PricelistHistoryDatabasesDir: config.DatabaseConfig.PricelistHistoryDir,
		Tuples:                       sta.RegionState.RegionComposites.ToTuples(),
		ReceiveRegionTimestamps:      sta.RegionState.ReceiveTimestamps,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise pricelist-history state")

		return ApiState{}, err
	}

	// establishing listeners
	sta.Listeners = state.NewListeners(state.NewSubjectListeners([]state.SubjectListeners{
		sta.ItemsState.GetListeners(),
		sta.AreaMapsState.GetListeners(),
		sta.TokensState.GetListeners(),
		sta.RegionState.GetListeners(),
		sta.BootState.GetListeners(),
		sta.LiveAuctionsState.GetListeners(),
		sta.PricelistHistoryState.GetListeners(),
	}))

	return sta, nil
}

type ApiState struct {
	state.State

	BlizzardState state.BlizzardState
	ItemsState    state.ItemsState
	AreaMapsState state.AreaMapsState
	TokensState   state.TokensState
	RegionState   *state.RegionsState
	Collector     BaseCollector.Client
	BootState     state.BootState

	LiveAuctionsState     state.LiveAuctionsState
	PricelistHistoryState state.PricelistHistoryState
}
