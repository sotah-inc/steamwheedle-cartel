package subjects

// Subject - typehint for these enums
type Subject string

/*
Status - subject name for returning current status
*/
const (
	Status                      Subject = "status"
	Auctions                    Subject = "auctions"
	GenericTestErrors           Subject = "genericTestErrors"
	ItemsQuery                  Subject = "itemsQuery"
	TokenHistory                Subject = "tokenHistory"
	ValidateRegionRealm         Subject = "validateRegionRealm"
	PriceList                   Subject = "priceList"
	PriceListHistory            Subject = "priceListHistory"
	Items                       Subject = "items"
	Boot                        Subject = "boot"
	SessionSecret               Subject = "sessionSecret"
	LiveAuctionsIntake          Subject = "liveAuctionsIntake"
	PricelistHistoriesIntake    Subject = "pricelistHistoriesIntake"
	AppMetrics                  Subject = "appMetrics"
	PricelistHistoriesIntakeV2  Subject = "pricelistHistoriesIntakeV2"
	QueryRealmModificationDates Subject = "queryRealmModificationDates"
	RealmModificationDates      Subject = "realmModificationDates"
	QueryAuctionStats           Subject = "queryAuctionStats"
	AreaMapsQuery               Subject = "areaMapsQuery"
	AreaMaps                    Subject = "areaMaps"
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
