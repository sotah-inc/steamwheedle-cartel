package run

import (
	"log"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah/gameversions"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store/regions"
	"github.com/twinj/uuid"
)

type SyncItemIconsStateConfig struct {
	ProjectId string
}

func NewSyncItemIconsState(config SyncItemIconsStateConfig) (SyncItemIconsState, error) {
	// establishing an initial state
	sta := SyncItemIconsState{
		State: state.NewState(uuid.NewV4(), true),
	}

	var err error

	// initializing a bus client
	sta.IO.BusClient, err = bus.NewClient(config.ProjectId, "run-sync-items")
	if err != nil {
		log.Fatalf("Failed to create new bus client: %s", err.Error())

		return SyncItemIconsState{}, err
	}
	sta.receiveSyncedItemsTopic, err = sta.IO.BusClient.FirmTopic(string(subjects.ReceiveSyncedItems))
	if err != nil {
		log.Fatalf("Failed to get firm topic: %s", err.Error())

		return SyncItemIconsState{}, err
	}

	// initializing a store client
	sta.IO.StoreClient, err = store.NewClient(config.ProjectId)
	if err != nil {
		log.Fatalf("Failed to create new store client: %s", err.Error())

		return SyncItemIconsState{}, err
	}

	sta.itemsBase = store.NewItemsBase(sta.IO.StoreClient, regions.USCentral1, gameversions.Retail)
	sta.itemsBucket, err = sta.itemsBase.GetFirmBucket()
	if err != nil {
		log.Fatalf("Failed to get firm bucket: %s", err.Error())

		return SyncItemIconsState{}, err
	}

	sta.itemIconsBase = store.NewItemIconsBase(sta.IO.StoreClient, regions.USCentral1, gameversions.Retail)
	sta.itemIconsBucket, err = sta.itemIconsBase.GetFirmBucket()
	if err != nil {
		log.Fatalf("Failed to get firm bucket: %s", err.Error())

		return SyncItemIconsState{}, err
	}

	// resolving item-icons bucket name
	bktAttrs, err := sta.itemIconsBucket.Attrs(sta.IO.StoreClient.Context)
	if err != nil {
		log.Fatalf("Failed to get bucket attrs: %s", err.Error())

		return SyncItemIconsState{}, err
	}
	sta.itemIconsBucketName = bktAttrs.Name

	return sta, nil
}

type SyncItemIconsState struct {
	state.State

	receiveSyncedItemsTopic *pubsub.Topic

	itemsBase   store.ItemsBase
	itemsBucket *storage.BucketHandle

	itemIconsBase       store.ItemIconsBase
	itemIconsBucket     *storage.BucketHandle
	itemIconsBucketName string
}
