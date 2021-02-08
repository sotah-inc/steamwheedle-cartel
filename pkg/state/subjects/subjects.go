package subjects

type Subject string

const (
	ItemsIntake Subject = "itemsIntake"
	Items       Subject = "items"
	ItemsQuery  Subject = "itemsQuery"
)

const (
	Pets       Subject = "pets"
	PetsIntake Subject = "petsIntake"
	PetsQuery  Subject = "petsQuery"
)

const (
	Professions             Subject = "professions"
	ProfessionsIntake       Subject = "professionsIntake"
	SkillTiersIntake        Subject = "skillTiersIntake"
	RecipesIntake           Subject = "recipesIntake"
	SkillTier               Subject = "skillTier"
	Recipe                  Subject = "recipe"
	MiniRecipes             Subject = "miniRecipes"
	PrunePricelistHistories Subject = "prunePricelistHistories"
)

const (
	AreaMapsQuery Subject = "areaMapsQuery"
	AreaMaps      Subject = "areaMaps"
)

const (
	Status                          Subject = "status"
	ConnectedRealms                 Subject = "connectedRealms"
	ValidateRegionConnectedRealm    Subject = "validateRegionConnectedRealm"
	ResolveConnectedRealm           Subject = "resolveConnectedRealm"
	ValidateRegionRealm             Subject = "validateRegionRealm"
	QueryRealmModificationDates     Subject = "queryRealmModificationDates"
	ConnectedRealmModificationDates Subject = "connectedRealmModificationDates"
	ReceiveRegionTimestamps         Subject = "receiveRegionTimestamps"
)

const (
	Boot          Subject = "boot"
	SessionSecret Subject = "sessionSecret"
)

const (
	TokenHistoryIntake Subject = "tokenHistoryIntake"
	RegionTokenHistory Subject = "regionTokenHistory"
	TokenHistory       Subject = "tokenHistory"
)

const (
	LiveAuctionsIntake Subject = "liveAuctionsIntake"
	Auctions           Subject = "auctions"
	PriceList          Subject = "priceList"
)

const (
	StatsIntake       Subject = "statsIntake"
	QueryAuctionStats Subject = "queryAuctionStats"
)

const (
	ItemPricesIntake    Subject = "itemPricesIntake"
	RecipePricesIntake  Subject = "recipePricesIntake"
	ItemPricesHistory   Subject = "itemPricesHistory"
	RecipePricesHistory Subject = "recipePricesHistory"
	ItemsMarketPrice    Subject = "itemsMarketPrice"
)
