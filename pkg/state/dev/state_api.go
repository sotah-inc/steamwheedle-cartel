package dev

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/twinj/uuid"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/bus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/diskstore"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/metric"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/resolver"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/store"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type APIStateConfig struct {
	SotahConfig sotah.Config

	GCloudProjectID string

	MessengerHost string
	MessengerPort int

	DiskStoreCacheDir string

	BlizzardClientId     string
	BlizzardClientSecret string

	ItemsDatabaseDir    string
	TokensDatabaseDir   string
	AreaMapsDatabaseDir string
}

func NewAPIState(config APIStateConfig) (*APIState, error) {
	// establishing an initial state
	apiState := APIState{
		State: state.NewState(uuid.NewV4(), config.SotahConfig.UseGCloud),
	}
	apiState.SessionSecret = uuid.NewV4()

	// setting api-state from config, including filtering in regions based on config whitelist
	apiState.Statuses = sotah.Statuses{}
	apiState.Regions = config.SotahConfig.FilterInRegions(config.SotahConfig.Regions)
	apiState.Expansions = config.SotahConfig.Expansions
	apiState.Professions = config.SotahConfig.Professions
	apiState.ItemBlacklist = config.SotahConfig.ItemBlacklist

	// establishing a store (gcloud store or disk store)
	if config.SotahConfig.UseGCloud {
		stor, err := store.NewClient(config.GCloudProjectID)
		if err != nil {
			return nil, err
		}

		apiState.IO.StoreClient = stor

		// establishing a bus
		logging.Info("Connecting bus-client")
		busClient, err := bus.NewClient(config.GCloudProjectID, "api")
		if err != nil {
			return nil, err
		}
		apiState.IO.BusClient = busClient
	} else {
		if config.DiskStoreCacheDir == "" {
			return nil, errors.New("disk-store-cache-dir should not be blank")
		}

		cacheDirs := []string{
			config.DiskStoreCacheDir,
			fmt.Sprintf("%s/items", config.DiskStoreCacheDir),
			fmt.Sprintf("%s/auctions", config.DiskStoreCacheDir),
			fmt.Sprintf("%s/databases", config.DiskStoreCacheDir),
		}
		for _, reg := range apiState.Regions {
			cacheDirs = append(cacheDirs, fmt.Sprintf("%s/auctions/%s", config.DiskStoreCacheDir, reg.Name))
		}
		if err := util.EnsureDirsExist(cacheDirs); err != nil {
			return nil, err
		}

		apiState.IO.DiskStore = diskstore.NewDiskStore(config.DiskStoreCacheDir)
	}

	// connecting to the messenger host
	mess, err := messenger.NewMessenger(config.MessengerHost, config.MessengerPort)
	if err != nil {
		return nil, err
	}
	apiState.IO.Messenger = mess

	// initializing a reporter
	apiState.IO.Reporter = metric.NewReporter(mess)

	// connecting a new blizzard client
	blizzardClient, err := blizzard.NewClient(config.BlizzardClientId, config.BlizzardClientSecret)
	if err != nil {
		return nil, err
	}
	apiState.IO.Resolver = resolver.NewResolver(blizzardClient, apiState.IO.Reporter)

	// filling state with region statuses
	for job := range apiState.IO.Resolver.GetStatuses(apiState.Regions) {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("failed to fetch status for region")

			return nil, job.Err
		}

		job.Status.Realms = config.SotahConfig.FilterInRealms(job.Region, job.Status.Realms)
		apiState.Statuses[job.Region.Name] = job.Status
	}

	// filling state with item-classes
	primaryRegion, err := apiState.Regions.GetPrimaryRegion()
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error":   err.Error(),
			"regions": apiState.Regions,
		}).Error("failed to retrieve primary region")

		return nil, err
	}
	uri, err := apiState.IO.Resolver.AppendAccessToken(apiState.IO.Resolver.GetItemClassesURL(primaryRegion.Hostname))
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error":                   err.Error(),
			"primary-region-hostname": primaryRegion.Hostname,
		}).Error("failed to append access-token to get-item-classes url")

		return nil, err
	}
	itemClasses, _, err := blizzard.NewItemClassesFromHTTP(uri)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   uri,
		}).Error("failed to get item-classes via http")

		return nil, err
	}
	apiState.ItemClasses = itemClasses

	// loading the items database
	itemsDatabase, err := database.NewItemsDatabase(config.ItemsDatabaseDir)
	if err != nil {
		return nil, err
	}
	apiState.IO.Databases.ItemsDatabase = itemsDatabase

	// loading the tokens database
	tokensDatabase, err := database.NewTokensDatabase(config.TokensDatabaseDir)
	if err != nil {
		return nil, err
	}
	apiState.IO.Databases.TokensDatabase = tokensDatabase

	// loading the area-maps database
	areaMapsDatabase, err := database.NewAreaMapsDatabase(config.AreaMapsDatabaseDir)
	if err != nil {
		return &APIState{}, err
	}
	apiState.AreaMapsDatabase = areaMapsDatabase

	// gathering profession icons
	for i, prof := range apiState.Professions {
		apiState.Professions[i].IconURL = blizzard.DefaultGetItemIconURL(prof.Icon)
	}

	// establishing listeners
	apiState.Listeners = state.NewListeners(state.SubjectListeners{
		subjects.Boot:                        apiState.ListenForBoot,
		subjects.SessionSecret:               apiState.ListenForSessionSecret,
		subjects.Status:                      apiState.ListenForStatus,
		subjects.Items:                       apiState.ListenForItems,
		subjects.ItemsQuery:                  apiState.ListenForItemsQuery,
		subjects.QueryRealmModificationDates: apiState.ListenForQueryRealmModificationDates,
		subjects.RealmModificationDates:      apiState.ListenForRealmModificationDates,
		subjects.TokenHistory:                apiState.ListenForTokenHistory,
		subjects.ValidateRegionRealm:         apiState.ListenForValidateRegionRealm,
		subjects.AreaMapsQuery:               apiState.ListenForAreaMapsQuery,
		subjects.AreaMaps:                    apiState.ListenForAreaMaps,
	})

	apiState.RegionRealmModificationDates = sotah.RegionRealmModificationDates{}

	return &apiState, nil
}

type APIState struct {
	state.State

	Regions  sotah.RegionList
	Statuses sotah.Statuses

	SessionSecret uuid.UUID
	ItemClasses   blizzard.ItemClasses
	Expansions    []sotah.Expansion
	Professions   []sotah.Profession
	ItemBlacklist state.ItemBlacklist

	RegionRealmModificationDates sotah.RegionRealmModificationDates

	AreaMapsDatabase database.AreaMapsDatabase
}
