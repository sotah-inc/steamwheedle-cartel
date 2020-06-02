package subjects

type Subject string

const (
	ItemsIntake Subject = "itemsIntake"
	Items       Subject = "items"
	ItemsQuery  Subject = "itemsQuery"
)

const (
	AreaMapsQuery Subject = "areaMapsQuery"
	AreaMaps      Subject = "areaMaps"
)

const (
	Status                          Subject = "status"
	ValidateRegionConnectedRealm    Subject = "validateRegionConnectedRealm"
	ValidateRegionRealm             Subject = "validateRegionRealm"
	QueryRealmModificationDates     Subject = "queryRealmModificationDates"
	ConnectedRealmModificationDates Subject = "connectedRealmModificationDates"
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
	PriceListHistory       Subject = "priceListHistory"
)
