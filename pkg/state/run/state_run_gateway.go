package run

import (
	"log"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/hell"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah/gameversions"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store/regions"
	"github.com/twinj/uuid"
)

type GatewayStateConfig struct {
	ProjectId string
}

func NewGatewayState(config GatewayStateConfig) (GatewayState, error) {
	// establishing an initial state
	sta := GatewayState{
		State: state.NewState(uuid.NewV4(), true),
	}

	var err error

	// connecting to hell
	sta.IO.HellClient, err = hell.NewClient(config.ProjectId)
	if err != nil {
		log.Fatalf("Failed to connect to firebase: %s", err.Error())

		return GatewayState{}, err
	}

	sta.actEndpoints, err = sta.IO.HellClient.GetActEndpoints()
	if err != nil {
		log.Fatalf("Failed to fetch act endpoints: %s", err.Error())

		return GatewayState{}, err
	}

	// initializing a bus client
	sta.IO.BusClient, err = bus.NewClient(config.ProjectId, "run-gateway")
	if err != nil {
		log.Fatalf("Failed to create new bus client: %s", err.Error())

		return GatewayState{}, err
	}
	sta.receiveRealmsTopic, err = sta.IO.BusClient.FirmTopic(string(subjects.ReceiveRealms))
	if err != nil {
		log.Fatalf("Failed to get firm topic: %s", err.Error())

		return GatewayState{}, err
	}
	sta.callComputeAllLiveAuctionsTopic, err = sta.IO.BusClient.FirmTopic(string(subjects.CallComputeAllLiveAuctions))
	if err != nil {
		log.Fatalf("Failed to get firm topic: %s", err.Error())

		return GatewayState{}, err
	}
	sta.receiveComputedLiveAuctionsTopic, err = sta.IO.BusClient.FirmTopic(string(subjects.ReceiveComputedLiveAuctions))
	if err != nil {
		log.Fatalf("Failed to get firm topic: %s", err.Error())

		return GatewayState{}, err
	}
	sta.filterInItemsToSyncTopic, err = sta.IO.BusClient.FirmTopic(string(subjects.FilterInItemsToSync))
	if err != nil {
		log.Fatalf("Failed to get firm topic: %s", err.Error())

		return GatewayState{}, err
	}
	sta.callSyncAllItemsTopic, err = sta.IO.BusClient.FirmTopic(string(subjects.CallSyncAllItems))
	if err != nil {
		log.Fatalf("Failed to get firm topic: %s", err.Error())

		return GatewayState{}, err
	}
	sta.callComputeAllPricelistHistoriesTopic, err = sta.IO.BusClient.FirmTopic(
		string(subjects.CallComputeAllPricelistHistories),
	)
	if err != nil {
		log.Fatalf("Failed to get firm topic: %s", err.Error())

		return GatewayState{}, err
	}
	sta.receiveComputedPricelistHistoriesTopic, err = sta.IO.BusClient.FirmTopic(
		string(subjects.ReceiveComputedPricelistHistories),
	)
	if err != nil {
		log.Fatalf("Failed to get firm topic: %s", err.Error())

		return GatewayState{}, err
	}

	// initializing a store client
	sta.IO.StoreClient, err = store.NewClient(config.ProjectId)
	if err != nil {
		log.Fatalf("Failed to create new store client: %s", err.Error())

		return GatewayState{}, err
	}

	sta.bootBase = store.NewBootBase(sta.IO.StoreClient, regions.USCentral1)
	sta.bootBucket, err = sta.bootBase.GetFirmBucket()
	if err != nil {
		log.Fatalf("Failed to get firm bucket: %s", err.Error())

		return GatewayState{}, err
	}

	sta.realmsBase = store.NewRealmsBase(sta.IO.StoreClient, regions.USCentral1, gameversions.Retail)
	sta.realmsBucket, err = sta.realmsBase.GetFirmBucket()
	if err != nil {
		log.Fatalf("Failed to get firm bucket: %s", err.Error())

		return GatewayState{}, err
	}

	return sta, nil
}

type GatewayState struct {
	state.State

	bootBase     store.BootBase
	bootBucket   *storage.BucketHandle
	realmsBase   store.RealmsBase
	realmsBucket *storage.BucketHandle

	actEndpoints hell.ActEndpoints

	receiveRealmsTopic                     *pubsub.Topic
	callComputeAllLiveAuctionsTopic        *pubsub.Topic
	receiveComputedLiveAuctionsTopic       *pubsub.Topic
	filterInItemsToSyncTopic               *pubsub.Topic
	callSyncAllItemsTopic                  *pubsub.Topic
	callComputeAllPricelistHistoriesTopic  *pubsub.Topic
	receiveComputedPricelistHistoriesTopic *pubsub.Topic
}
