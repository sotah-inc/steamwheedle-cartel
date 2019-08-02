package subjects

// Subject - typehint for these enums
type Subject string

/*
Status - subject name for returning current status
*/
const (
	Status                          Subject = "status"
	Auctions                        Subject = "auctions"
	GenericTestErrors               Subject = "genericTestErrors"
	Owners                          Subject = "owners"
	OwnersQueryByItems              Subject = "ownersQueryByItems"
	OwnersQuery                     Subject = "ownersQuery"
	ItemsQuery                      Subject = "itemsQuery"
	PriceList                       Subject = "priceList"
	PriceListHistory                Subject = "priceListHistory"
	PriceListHistoryV2              Subject = "priceListHistoryV2"
	Items                           Subject = "items"
	Boot                            Subject = "boot"
	SessionSecret                   Subject = "sessionSecret"
	RuntimeInfo                     Subject = "runtimeInfo"
	LiveAuctionsIntake              Subject = "liveAuctionsIntake"
	LiveAuctionsCompute             Subject = "liveAuctionsCompute"
	PricelistHistoriesIntake        Subject = "pricelistHistoriesIntake"
	AppMetrics                      Subject = "appMetrics"
	PricelistHistoriesIntakeV2      Subject = "pricelistHistoriesIntakeV2"
	PricelistHistoriesCompute       Subject = "pricelistHistoriesCompute"
	PricelistHistoriesComputeIntake Subject = "pricelistHistoriesComputeIntake"
	AuctionsQuery                   Subject = "auctionsQuery"
	QueryRealmModificationDates     Subject = "queryRealmModificationDates"
	RealmModificationDates          Subject = "realmModificationDates"
)

// gcloud fn-related
const (
	ReceiveComputedLiveAuctions Subject = "receiveComputedLiveAuctions"

	ComputeAllPricelistHistories      Subject = "computeAllPricelistHistories"
	ComputePricelistHistories         Subject = "computePricelistHistories"
	ReceiveComputedPricelistHistories Subject = "receiveComputedPricelistHistories"

	ReceiveRealms Subject = "receiveRealms"

	FilterInItemsToSync Subject = "filterInItemsToSync"
	SyncItemIcons       Subject = "syncItemIcons"
	ReceiveSyncedItems  Subject = "receiveSyncedItems"

	CleanupAllPricelistHistories Subject = "cleanupAllPricelistHistories"
	CleanupPricelistHistories    Subject = "cleanupPricelistHistories"
)

// gcloud gateway subjects
const (
	CallDownloadAllAuctions    Subject = "callDownloadAllAuctions"
	CallCleanupAllManifests    Subject = "callCleanupAllManifests"
	CallCleanupAllAuctions     Subject = "callCleanupAllAuctions"
	CallComputeAllLiveAuctions Subject = "callComputeAllLiveAuctions"
	CallSyncAllItems           Subject = "callSyncAllItems"
)
