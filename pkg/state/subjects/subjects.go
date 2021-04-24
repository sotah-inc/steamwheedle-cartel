package subjects

type Subject string

const (
	ItemsIntake              Subject = "itemsIntake"
	Items                    Subject = "items"
	ItemsQuery               Subject = "itemsQuery"
	ItemsFindMatchingRecipes Subject = "itemsFindMatchingRecipes"
	ItemClasses              Subject = "itemClasses"
	ItemSubjectsByItemClass  Subject = "itemSubjectsByItemClass"
)

const (
	Pets       Subject = "pets"
	PetsIntake Subject = "petsIntake"
	PetsQuery  Subject = "petsQuery"
)

const (
	Professions              Subject = "professions"
	ProfessionsFromIds       Subject = "professionsFromIds"
	ProfessionsIntake        Subject = "professionsIntake"
	SkillTiersIntake         Subject = "skillTiersIntake"
	SkillTiers               Subject = "skillTiers"
	RecipesIntake            Subject = "recipesIntake"
	ItemRecipesIntake        Subject = "itemRecipesIntake"
	SkillTier                Subject = "skillTier"
	Recipe                   Subject = "recipe"
	Recipes                  Subject = "recipes"
	RecipesQuery             Subject = "recipesQuery"
	MiniRecipes              Subject = "miniRecipes"
	ItemsRecipes             Subject = "itemsRecipes"
	ProfessionRecipeSubjects Subject = "professionRecipeSubjects"
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
	ItemPricesIntake        Subject = "itemPricesIntake"
	RecipePricesIntake      Subject = "recipePricesIntake"
	ItemPricesHistory       Subject = "itemPricesHistory"
	RecipePricesHistory     Subject = "recipePricesHistory"
	ItemsMarketPrice        Subject = "itemsMarketPrice"
	PrunePricelistHistories Subject = "prunePricelistHistories"
)
