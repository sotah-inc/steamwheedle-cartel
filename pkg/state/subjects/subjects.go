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
	PriceList          Subject = "priceList"
)

const (
	PricelistHistoryIntake Subject = "pricelistHistoryIntake"
)

const (
	GenericTestErrors          Subject = "genericTestErrors"
	PriceListHistory           Subject = "priceListHistory"
	AppMetrics                 Subject = "appMetrics"
	PricelistHistoriesIntakeV2 Subject = "pricelistHistoriesIntakeV2"
	RealmModificationDates     Subject = "realmModificationDates"
)

const (
	ReceiveRealms                     Subject = "receiveRealms"
	ReceiveComputedLiveAuctions       Subject = "receiveComputedLiveAuctions"
	ReceiveComputedPricelistHistories Subject = "receiveComputedPricelistHistories"

	FilterInItemsToSync Subject = "filterInItemsToSync"
	ReceiveSyncedItems  Subject = "receiveSyncedItems"

	SyncPubsubTopicsMonitor Subject = "syncPubsubTopicsMonitor"
)

const (
	CallDownloadAllAuctions          Subject = "callDownloadAllAuctions"
	CallCleanupAllManifests          Subject = "callCleanupAllManifests"
	CallCleanupAllAuctions           Subject = "callCleanupAllAuctions"
	CallComputeAllLiveAuctions       Subject = "callComputeAllLiveAuctions"
	CallSyncAllItems                 Subject = "callSyncAllItems"
	CallComputeAllPricelistHistories Subject = "callComputeAllPricelistHistories"
	CallCleanupAllPricelistHistories Subject = "callCleanupAllPricelistHistories"
)
