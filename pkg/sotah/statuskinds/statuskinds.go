package statuskinds

type StatusKind string

const (
	Downloaded           StatusKind = "downloaded"
	LiveAuctionsReceived StatusKind = "liveAuctionsReceived"
	ItemPricesReceived   StatusKind = "itemPricesReceived"
	RecipePricesReceived StatusKind = "recipePricesReceived"
	StatsReceived        StatusKind = "statsReceived"
)
