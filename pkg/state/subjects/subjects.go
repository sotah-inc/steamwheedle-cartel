package subjects

type Subject string

const (
	Items      Subject = "items"
	ItemsQuery Subject = "itemsQuery"
)

const (
	AreaMapsQuery Subject = "areaMapsQuery"
	AreaMaps      Subject = "areaMaps"
)

const (
	Status                      Subject = "status"
	ValidateRegionRealm         Subject = "validateRegionRealm"
	QueryRealmModificationDates Subject = "queryRealmModificationDates"
)

const (
	Boot          Subject = "boot"
	SessionSecret Subject = "sessionSecret"
)

const (
	TokenHistory Subject = "tokenHistory"
)

const (
	LiveAuctionsIntake Subject = "liveAuctionsIntake"
	Auctions           Subject = "auctions"
	QueryAuctionStats  Subject = "queryAuctionStats"
)

/*
Status - subject name for returning current status
*/
const (
	GenericTestErrors          Subject = "genericTestErrors"
	PriceList                  Subject = "priceList"
	PriceListHistory           Subject = "priceListHistory"
	PricelistHistoriesIntake   Subject = "pricelistHistoriesIntake"
	AppMetrics                 Subject = "appMetrics"
	PricelistHistoriesIntakeV2 Subject = "pricelistHistoriesIntakeV2"
	RealmModificationDates     Subject = "realmModificationDates"
)

// gcloud fn-related
const (
	ReceiveRealms                     Subject = "receiveRealms"
	ReceiveComputedLiveAuctions       Subject = "receiveComputedLiveAuctions"
	ReceiveComputedPricelistHistories Subject = "receiveComputedPricelistHistories"

	FilterInItemsToSync Subject = "filterInItemsToSync"
	ReceiveSyncedItems  Subject = "receiveSyncedItems"

	SyncPubsubTopicsMonitor Subject = "syncPubsubTopicsMonitor"
)

// gcloud gateway subjects
const (
	CallDownloadAllAuctions          Subject = "callDownloadAllAuctions"
	CallCleanupAllManifests          Subject = "callCleanupAllManifests"
	CallCleanupAllAuctions           Subject = "callCleanupAllAuctions"
	CallComputeAllLiveAuctions       Subject = "callComputeAllLiveAuctions"
	CallSyncAllItems                 Subject = "callSyncAllItems"
	CallComputeAllPricelistHistories Subject = "callComputeAllPricelistHistories"
	CallCleanupAllPricelistHistories Subject = "callCleanupAllPricelistHistories"
)
