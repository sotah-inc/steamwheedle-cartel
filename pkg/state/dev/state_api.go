package dev

import (
	"github.com/twinj/uuid"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseCollector "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/collector/base" // nolint:lll
	DiskCollector "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/collector/disk" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
)

type ApiStateDatabaseConfig struct {
	ItemsDir            string
	PetsDir             string
	ProfessionsDir      string
	TokensDir           string
	AreaMapsDir         string
	LiveAuctionsDir     string
	PricelistHistoryDir string
	StatsDir            string
}

type ApiStateConfig struct {
	SotahConfig       sotah.Config
	MessengerConfig   messenger.Config
	DiskStoreCacheDir string
	BlizzardConfig    blizzardv2.ClientConfig
	DatabaseConfig    ApiStateDatabaseConfig
	UseGCloud         bool
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

	// connecting to the messenger host
	logging.Info("producing new messenger")
	mess, err := messenger.NewMessenger(config.MessengerConfig)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to connect to messenger")

		return ApiState{}, err
	}

	// connecting a new blizzard client
	logging.Info("producing new blizzard-state")
	sta.BlizzardState, err = state.NewBlizzardState(config.BlizzardConfig)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise blizzard-client")

		return ApiState{}, err
	}

	// deriving lake-client
	logging.Info("producing new lake-client")
	lakeClient, err := lake.NewClient(lake.NewClientOptions{
		UseGCloud:   config.UseGCloud,
		CacheDir:    config.DiskStoreCacheDir,
		RegionNames: regions.Names(),
		ResolveItems: func(ids blizzardv2.ItemIds) chan blizzardv2.GetItemsOutJob {
			return sta.BlizzardState.ResolveItems(primaryRegion, ids)
		},
		ResolveItemMedias: sta.BlizzardState.ResolveItemMedias,
		ResolvePets: func(blacklist []blizzardv2.PetId) (chan blizzardv2.GetAllPetsJob, error) {
			return sta.BlizzardState.ResolvePets(primaryRegion, blacklist)
		},
		ResolveProfessions: func(
			blacklist []blizzardv2.ProfessionId,
		) (chan blizzardv2.GetAllProfessionsJob, error) {
			return sta.BlizzardState.ResolveProfessions(primaryRegion, blacklist)
		},
		ResolveProfessionMedias: sta.BlizzardState.ResolveProfessionMedias,
		ResolveSkillTiers: func(
			professionId blizzardv2.ProfessionId,
			idList []blizzardv2.SkillTierId,
		) chan blizzardv2.GetAllSkillTiersJob {
			return sta.BlizzardState.ResolveSkillTiers(primaryRegion, professionId, idList)
		},
		ResolveRecipes: func(ids []blizzardv2.RecipeId) chan blizzardv2.GetRecipesJob {
			return sta.BlizzardState.ResolveRecipes(primaryRegion, ids)
		},
		ResolveRecipeMedias: sta.BlizzardState.ResolveRecipeMedias,
		PrimarySkillTiers:   config.SotahConfig.PrimarySkillTiers,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise lake-client")

		return ApiState{}, err
	}

	// gathering region state
	logging.Info("producing new region-state")
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
	logging.Info("producing new boot-state")
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

	// resolving items state
	logging.Info("producing new items-state")
	sta.ItemsState, err = state.NewItemsState(state.NewItemsStateOptions{
		LakeClient:       lakeClient,
		Messenger:        mess,
		ItemsDatabaseDir: config.DatabaseConfig.ItemsDir,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise items state")

		return ApiState{}, err
	}

	// resolving pets state
	logging.Info("producing new pets-state")
	sta.PetsState, err = state.NewPetsState(state.NewPetsStateOptions{
		LakeClient:      lakeClient,
		Messenger:       mess,
		PetsDatabaseDir: config.DatabaseConfig.PetsDir,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise pets state")

		return ApiState{}, err
	}

	// resolving professions state
	logging.Info("producing new professions-state")
	sta.ProfessionsState, err = state.NewProfessionsState(state.NewProfessionsStateOptions{
		LakeClient:             lakeClient,
		Messenger:              mess,
		ProfessionsDatabaseDir: config.DatabaseConfig.ProfessionsDir,
		ProfessionsBlacklist:   config.SotahConfig.ProfessionsBlacklist,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise professions state")

		return ApiState{}, err
	}

	// loading the tokens state
	logging.Info("producing new tokens-state")
	sta.TokensState, err = state.NewTokensState(state.NewTokensStateOptions{
		BlizzardState:     sta.BlizzardState,
		Messenger:         mess,
		TokensDatabaseDir: config.DatabaseConfig.TokensDir,
		Regions:           regions,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise tokens state")

		return ApiState{}, err
	}

	// loading the area-maps state
	logging.Info("producing new area-maps state")
	sta.AreaMapsState, err = state.NewAreaMapsState(mess, config.DatabaseConfig.AreaMapsDir)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise area-maps state")

		return ApiState{}, err
	}

	// resolving disk-collector-auctions state
	logging.Info("producing new disk-collector client")
	sta.Collector = DiskCollector.NewClient(DiskCollector.ClientOptions{
		ResolveAuctions: func() chan blizzardv2.GetAuctionsJob {
			return sta.BlizzardState.ResolveAuctions(sta.RegionState.RegionComposites.ToDownloadTuples())
		},
		ReceiveRegionTimestamps: sta.RegionState.ReceiveTimestamps,
		LakeClient:              lakeClient,
		MessengerClient:         mess,
	})

	// resolving live-auctions state
	logging.Info("producing new live-auctions state")
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
	logging.Info("producing new pricelist-history state")
	sta.PricelistHistoryState, err = state.NewPricelistHistoryState(
		state.NewPricelistHistoryStateOptions{
			Messenger:                    mess,
			LakeClient:                   lakeClient,
			PricelistHistoryDatabasesDir: config.DatabaseConfig.PricelistHistoryDir,
			Tuples:                       sta.RegionState.RegionComposites.ToTuples(),
			ReceiveRegionTimestamps:      sta.RegionState.ReceiveTimestamps,
		},
	)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise pricelist-history state")

		return ApiState{}, err
	}

	// resolving stats state
	logging.Info("producing new stats state")
	sta.StatsState, err = state.NewStatsState(state.NewStatsStateOptions{
		Messenger:               mess,
		LakeClient:              lakeClient,
		StatsDatabasesDir:       config.DatabaseConfig.StatsDir,
		Tuples:                  sta.RegionState.RegionComposites.ToTuples(),
		ReceiveRegionTimestamps: sta.RegionState.ReceiveTimestamps,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise stats state")

		return ApiState{}, err
	}

	// establishing listeners
	logging.Info("establishing listeners")
	sta.Listeners = state.NewListeners(state.NewSubjectListeners([]state.SubjectListeners{
		sta.ItemsState.GetListeners(),
		sta.PetsState.GetListeners(),
		sta.ProfessionsState.GetListeners(),
		sta.AreaMapsState.GetListeners(),
		sta.TokensState.GetListeners(),
		sta.RegionState.GetListeners(),
		sta.BootState.GetListeners(),
		sta.LiveAuctionsState.GetListeners(),
		sta.PricelistHistoryState.GetListeners(),
		sta.StatsState.GetListeners(),
	}))

	return sta, nil
}

type ApiState struct {
	state.State

	BlizzardState    state.BlizzardState
	ItemsState       state.ItemsState
	PetsState        state.PetsState
	ProfessionsState state.ProfessionsState
	AreaMapsState    state.AreaMapsState
	TokensState      state.TokensState
	RegionState      state.RegionsState
	Collector        BaseCollector.Client
	BootState        state.BootState

	LiveAuctionsState     state.LiveAuctionsState
	PricelistHistoryState state.PricelistHistoryState
	StatsState            state.StatsState
}
