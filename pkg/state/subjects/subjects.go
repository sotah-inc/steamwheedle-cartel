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
	Professions       Subject = "professions"
	ProfessionsIntake Subject = "professionsIntake"
	SkillTiersIntake  Subject = "skillTiersIntake"
	RecipesIntake     Subject = "recipesIntake"
	SkillTier         Subject = "skillTier"
	Recipe            Subject = "recipe"
	MiniRecipes       Subject = "miniRecipes"
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
	ItemPricesIntake Subject = "itemPricesIntake"
	PriceListHistory Subject = "priceListHistory"
)
