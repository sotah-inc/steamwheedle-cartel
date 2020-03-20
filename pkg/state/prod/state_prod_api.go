package prod

import (
	"fmt"
	"log"

	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"github.com/twinj/uuid"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/bus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/hell"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/metric"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/resolver"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/store"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/store/regions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type ProdApiStateConfig struct {
	SotahConfig sotah.Config

	GCloudProjectID string

	MessengerHost string
	MessengerPort int
}

func NewProdApiState(config ProdApiStateConfig) (ApiState, error) {
	// establishing an initial state
	apiState := ApiState{
		State: state.NewState(uuid.NewV4(), config.SotahConfig.UseGCloud),
	}
	apiState.SessionSecret = uuid.NewV4()

	// setting api-state from config, including filtering in regions based on config whitelist
	apiState.Statuses = sotah.Statuses{}
	apiState.Regions = config.SotahConfig.FilterInRegions(config.SotahConfig.Regions)
	apiState.Expansions = config.SotahConfig.Expansions
	apiState.Professions = config.SotahConfig.Professions

	// establishing a store
	stor, err := store.NewClient(config.GCloudProjectID)
	if err != nil {
		return ApiState{}, err
	}
	apiState.IO.StoreClient = stor

	bootBase := store.NewBootBase(apiState.IO.StoreClient, regions.USCentral1)

	var bootBucket *storage.BucketHandle
	bootBucket, err = bootBase.GetFirmBucket()
	if err != nil {
		return ApiState{}, err
	}
	blizzardCredentials, err := bootBase.GetBlizzardCredentials(bootBucket)
	if err != nil {
		return ApiState{}, err
	}

	apiState.RealmsBase = store.NewRealmsBase(apiState.IO.StoreClient, regions.USCentral1, gameversions.Retail)
	apiState.RealmsBucket, err = apiState.RealmsBase.GetFirmBucket()
	if err != nil {
		return ApiState{}, err
	}

	// establishing a bus
	logging.Info("Connecting bus-client")
	busClient, err := bus.NewClient(config.GCloudProjectID, "prod-api")
	if err != nil {
		return ApiState{}, err
	}
	apiState.IO.BusClient = busClient

	// connecting to hell
	apiState.IO.HellClient, err = hell.NewClient(config.GCloudProjectID)
	if err != nil {
		log.Fatalf("Failed to connect to firebase: %s", err.Error())

		return ApiState{}, err
	}

	// connecting to the messenger host
	mess, err := messenger.NewMessenger(config.MessengerHost, config.MessengerPort)
	if err != nil {
		return ApiState{}, err
	}
	apiState.IO.Messenger = mess

	// initializing a reporter
	apiState.IO.Reporter = metric.NewReporter(mess)

	// connecting a new blizzard client
	blizzardClient, err := blizzard.NewClient(blizzardCredentials.ClientId, blizzardCredentials.ClientSecret)
	if err != nil {
		return ApiState{}, err
	}
	apiState.IO.Resolver = resolver.NewResolver(blizzardClient, apiState.IO.Reporter)

	// filling state with region statuses
	for _, region := range apiState.Regions {
		realms, err := apiState.RealmsBase.GetAllRealms(region.Name, apiState.RealmsBucket)
		if err != nil {
			return ApiState{}, err
		}

		status := apiState.Statuses[region.Name]
		status.Realms = config.SotahConfig.FilterInRealms(region, realms)
		apiState.Statuses[region.Name] = status
	}

	// filling state with item-classes
	primaryRegion, err := apiState.Regions.GetPrimaryRegion()
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error":   err.Error(),
			"regions": apiState.Regions,
		}).Error("Failed to retrieve primary region")

		return ApiState{}, err
	}
	itemClasses, err := apiState.ResolveItemClasses(primaryRegion.Hostname)
	if err != nil {
		return ApiState{}, err
	}
	apiState.ItemClasses = itemClasses

	// gathering profession icons
	itemIconsBase := store.NewItemIconsBase(stor, regions.USCentral1, gameversions.Retail)
	itemIconsBucket, err := itemIconsBase.GetFirmBucket()
	if err != nil {
		return ApiState{}, err
	}
	for i, prof := range apiState.Professions {
		itemIconUrl, err := func() (string, error) {
			obj := itemIconsBase.GetObject(prof.Icon, itemIconsBucket)
			exists, err := itemIconsBase.ObjectExists(obj)
			if err != nil {
				return "", err
			}

			url := fmt.Sprintf(
				store.ItemIconURLFormat,
				itemIconsBase.GetBucketName(),
				itemIconsBase.GetObjectName(prof.Icon),
			)

			if exists {
				return url, nil
			}

			body, err := util.Download(blizzard.DefaultGetItemIconURL(prof.Icon))
			if err != nil {
				return "", err
			}

			if err := itemIconsBase.Write(obj.NewWriter(stor.Context), body); err != nil {
				return "", err
			}

			return url, nil
		}()
		if err != nil {
			return ApiState{}, err
		}

		apiState.Professions[i].IconURL = itemIconUrl
	}

	apiState.HellRegionRealms, err = func() (hell.RegionRealmsMap, error) {
		out := hell.RegionRealmsMap{}

		hellRegionRealms, err := apiState.IO.HellClient.GetRegionRealms(
			apiState.Statuses.RegionRealmsMap().ToRegionRealmSlugs(),
			gameversions.Retail,
		)
		if err != nil {
			return hell.RegionRealmsMap{}, err
		}

		return out.Merge(hellRegionRealms), nil
	}()
	if err != nil {
		return ApiState{}, err
	}

	// establishing bus-listeners
	apiState.BusListeners = state.NewBusListeners(state.SubjectBusListeners{
		subjects.Status:        apiState.ListenForBusStatus,
		subjects.ReceiveRealms: apiState.ListenForReceiveRealms,
	})

	// establishing messenger-listeners
	apiState.Listeners = state.NewListeners(state.SubjectListeners{
		subjects.Boot:                        apiState.ListenForMessengerBoot,
		subjects.Status:                      apiState.ListenForMessengerStatus,
		subjects.SessionSecret:               apiState.ListenForSessionSecret,
		subjects.QueryRealmModificationDates: apiState.ListenForQueryRealmModificationDates,
		subjects.RealmModificationDates:      apiState.ListenForRealmModificationDates,
	})

	return apiState, nil
}

type ApiState struct {
	state.State

	Regions  sotah.RegionList
	Statuses sotah.Statuses

	RealmsBase   store.RealmsBase
	RealmsBucket *storage.BucketHandle

	SessionSecret uuid.UUID
	Expansions    []sotah.Expansion
	ItemClasses   sotah.ItemClasses
	Professions   []sotah.Profession
	ItemBlacklist state.ItemBlacklist

	BlizzardClientId     string
	BlizzardClientSecret string

	HellRegionRealms hell.RegionRealmsMap
}
