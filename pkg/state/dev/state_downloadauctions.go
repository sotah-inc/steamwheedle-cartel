package dev

import (
	"github.com/twinj/uuid"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	BaseCollector "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/collector/base" // nolint:lll
	DiskCollector "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/collector/disk" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/featureflags"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
)

type DownloadAuctionsDatabaseConfig struct {
	RegionsDir      string
	LiveAuctionsDir string
}

type DownloadAuctionsStateConfig struct {
	SotahConfig       sotah.Config
	MessengerConfig   messenger.Config
	DiskStoreCacheDir string
	BlizzardConfig    blizzardv2.ClientConfig
	DatabaseConfig    DownloadAuctionsDatabaseConfig
	UseGCloud         bool
}

func NewDownloadAuctionsState(config DownloadAuctionsStateConfig) (DownloadAuctionsState, error) {
	logging.Info("starting download-auctions state")

	// establishing an initial state
	sta := DownloadAuctionsState{
		State: state.State{RunID: uuid.NewV4(), Listeners: nil, BusListeners: nil},
	}

	// narrowing regions list
	regions := config.SotahConfig.FilterInRegions(config.SotahConfig.Regions)

	// resolving primary-region
	primaryRegion, err := regions.GetPrimaryRegion()
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to resolve primary-region")

		return DownloadAuctionsState{}, err
	}

	// connecting to the messenger host
	logging.Info("producing new messenger")
	mess, err := messenger.NewMessenger(config.MessengerConfig)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to connect to messenger")

		return DownloadAuctionsState{}, err
	}

	// connecting a new blizzard client
	logging.Info("producing new blizzard-state")
	blizzardState, err := state.NewBlizzardState(config.BlizzardConfig)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise blizzard-client")

		return DownloadAuctionsState{}, err
	}

	// deriving lake-client
	logging.Info("producing new lake-client")
	lakeClient, err := lake.NewClient(lake.NewClientOptions{
		UseGCloud:    config.UseGCloud,
		CacheDir:     config.DiskStoreCacheDir,
		RegionNames:  regions.Names(),
		GameVersions: config.SotahConfig.GameVersions(),
		ResolveItems: func(
			version gameversion.GameVersion,
			ids blizzardv2.ItemIds,
		) chan blizzardv2.GetItemsOutJob {
			return blizzardState.ResolveItems(primaryRegion, version, ids)
		},
		ResolveItemMedias: blizzardState.ResolveItemMedias,
		ResolvePets: func(blacklist []blizzardv2.PetId) (chan blizzardv2.GetAllPetsJob, error) {
			return blizzardState.ResolvePets(primaryRegion, blacklist)
		},
		ResolveProfessions: func(
			blacklist []blizzardv2.ProfessionId,
		) (chan blizzardv2.GetAllProfessionsJob, error) {
			return blizzardState.ResolveProfessions(primaryRegion, blacklist)
		},
		ResolveProfessionMedias: blizzardState.ResolveProfessionMedias,
		ResolveSkillTiers: func(
			professionId blizzardv2.ProfessionId,
			idList []blizzardv2.SkillTierId,
		) chan blizzardv2.GetAllSkillTiersJob {
			return blizzardState.ResolveSkillTiers(primaryRegion, professionId, idList)
		},
		ResolveRecipes: func(group blizzardv2.RecipesGroup) chan blizzardv2.GetRecipesOutJob {
			return blizzardState.ResolveRecipes(primaryRegion, group)
		},
		ResolveRecipeMedias: blizzardState.ResolveRecipeMedias,
		PrimarySkillTiers:   config.SotahConfig.PrimarySkillTiers,
		ResolveItemClasses: func() ([]blizzardv2.ItemClassResponse, error) {
			return blizzardState.ResolveItemClasses(regions)
		},
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise lake-client")

		return DownloadAuctionsState{}, err
	}

	// gathering region state
	logging.Info("producing new region-state")
	regionState, err := state.NewRegionState(state.NewRegionStateOptions{
		BlizzardState:      blizzardState,
		Regions:            regions,
		Messenger:          mess,
		RealmSlugWhitelist: config.SotahConfig.Whitelist,
		RegionsDatabaseDir: config.DatabaseConfig.RegionsDir,
		GameVersionList:    config.SotahConfig.GameVersions(),
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to establish region state")

		return DownloadAuctionsState{}, err
	}

	// resolving disk-collector-auctions state
	logging.Info("producing new disk-collector client")
	sta.Collector = DiskCollector.NewClient(DiskCollector.ClientOptions{
		ResolveAuctions: func() (chan blizzardv2.GetAuctionsJob, error) {
			downloadTuples, err := regionState.ResolveDownloadTuples()
			if err != nil {
				return nil, err
			}

			var nextTuples []blizzardv2.DownloadConnectedRealmTuple
			for _, tuple := range downloadTuples {
				isFetchAuctionsEnabled := func() bool {
					for _, meta := range config.SotahConfig.GameVersionMeta {
						if meta.Name != tuple.Version {
							continue
						}

						for _, flag := range meta.FeatureFlags {
							if flag == featureflags.Auctions {
								return true
							}
						}
					}

					return false
				}()
				if !isFetchAuctionsEnabled {
					continue
				}

				nextTuples = append(nextTuples, tuple)
			}

			return blizzardState.ResolveAuctions(nextTuples), nil
		},
		ReceiveRegionTimestamps: regionState.ReceiveTimestamps,
		LakeClient:              lakeClient,
		MessengerClient:         mess,
	})

	// resolving all tuples
	tuples, err := regionState.ResolveTuples()
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to resolve all tuples")

		return DownloadAuctionsState{}, err
	}

	// resolving live-auctions state
	logging.Info("producing new live-auctions state")
	sta.LiveAuctionsState, err = state.NewLiveAuctionsState(state.NewLiveAuctionsStateOptions{
		Messenger:                mess,
		LakeClient:               lakeClient,
		LiveAuctionsDatabasesDir: config.DatabaseConfig.LiveAuctionsDir,
		ReceiveRegionTimestamps:  regionState.ReceiveTimestamps,
		Tuples:                   tuples,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise live-auctions state")

		return DownloadAuctionsState{}, err
	}

	// establishing listeners
	logging.Info("establishing listeners")
	sta.Listeners = state.NewListeners(state.NewSubjectListeners([]state.SubjectListeners{
		sta.LiveAuctionsState.GetListeners(),
	}))

	return sta, nil
}

type DownloadAuctionsState struct {
	state.State

	LiveAuctionsState state.LiveAuctionsState
	Collector         BaseCollector.Client
}
