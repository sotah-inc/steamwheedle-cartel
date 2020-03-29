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
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
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
	apiState := APIState{
		State:         state.NewState(uuid.NewV4(), false),
		sessionSecret: uuid.NewV4(),
		regions:       config.SotahConfig.FilterInRegions(config.SotahConfig.Regions),
		expansions:    config.SotahConfig.Expansions,
		professions:   config.SotahConfig.Professions,
		itemBlacklist: config.SotahConfig.ItemBlacklist,
	}

	var err error
	apiState.diskStore, err = func() (diskstore.DiskStore, error) {
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
		for _, reg := range apiState.regions {
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
	apiState.messenger, err = messenger.NewMessenger(config.MessengerConfig.Hostname, config.MessengerConfig.Port)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to connect to messenger")

		return nil, err
	}

	// initializing a reporter
	apiState.reporter = metric.NewReporter(apiState.messenger)

	// connecting a new blizzard client
	apiState.blizzardClient, err = blizzardv2.NewClient(config.BlizzardConfig.ClientId, config.BlizzardConfig.ClientSecret)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise blizzard-client")

		return nil, err
	}

	// gathering connected-realms
	apiState.connectedRealms, err = func() (blizzardv2.ConnectedRealmResponses, error) {
		primaryRegion, err := apiState.regions.GetPrimaryRegion()
		if err != nil {
			return blizzardv2.ConnectedRealmResponses{}, err
		}

		return blizzardv2.GetAllConnectedRealms(blizzardv2.GetAllConnectedRealmsOptions{
			GetConnectedRealmIndexURL: func() (string, error) {
				return apiState.blizzardClient.AppendAccessToken(
					blizzardv2.DefaultConnectedRealmIndexURL(primaryRegion.Hostname, primaryRegion.Name),
				)
			},
			GetConnectedRealmURL: apiState.blizzardClient.AppendAccessToken,
		})
	}()
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to get connected-realms")

		return nil, err
	}

	// gathering item-classes
	apiState.itemClasses, err = func() ([]blizzardv2.ItemClassResponse, error) {
		primaryRegion, err := apiState.regions.GetPrimaryRegion()
		if err != nil {
			return []blizzardv2.ItemClassResponse{}, err
		}

		return blizzardv2.GetAllItemClasses(blizzardv2.GetAllItemClassesOptions{
			GetItemClassIndexURL: func() (string, error) {
				return apiState.blizzardClient.AppendAccessToken(
					blizzardv2.DefaultGetItemClassIndexURL(primaryRegion.Hostname, primaryRegion.Name),
				)
			},
			GetItemClassURL: func(id blizzardv2.ItemClassId) (string, error) {
				return apiState.blizzardClient.AppendAccessToken(
					blizzardv2.DefaultGetItemClassURL(primaryRegion.Hostname, primaryRegion.Name, id),
				)
			},
		})
	}()
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to get item-classes")

		return nil, err
	}

	// loading the items database
	apiState.itemsDatabase, err = database.NewItemsDatabase(config.DatabaseConfig.ItemsDir)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise items-database")

		return nil, err
	}

	// loading the tokens database
	apiState.tokensDatabase, err = database.NewTokensDatabase(config.DatabaseConfig.TokensDir)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise tokens-database")

		return nil, err
	}

	// loading the area-maps database
	apiState.areaMapsDatabase, err = database.NewAreaMapsDatabase(config.DatabaseConfig.AreaMapsDir)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise area-maps-database")

		return nil, err
	}

	// gathering profession icons
	for i, prof := range apiState.professions {
		apiState.professions[i].IconURL = blizzardv2.DefaultGetItemIconURL(prof.Icon)
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

	return &apiState, nil
}

type APIState struct {
	state.State

	// set at run-time
	sessionSecret   uuid.UUID
	connectedRealms blizzardv2.ConnectedRealmResponses
	itemClasses     []blizzardv2.ItemClassResponse

	// derived from config file
	regions       sotah.RegionList
	expansions    []sotah.Expansion
	professions   []sotah.Profession
	itemBlacklist state.ItemBlacklist

	// io
	areaMapsDatabase database.AreaMapsDatabase
	itemsDatabase    database.ItemsDatabase
	tokensDatabase   database.TokensDatabase
	diskStore        diskstore.DiskStore
	messenger        messenger.Messenger
	reporter         metric.Reporter
	blizzardClient   blizzardv2.Client
}
