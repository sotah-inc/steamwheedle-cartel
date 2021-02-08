package sotah

type ConnectedRealmTimestamps struct {
	Downloaded           UnixTimestamp `json:"downloaded"`
	LiveAuctionsReceived UnixTimestamp `json:"live_auctions_received"`
	ItemPricesReceived   UnixTimestamp `json:"item_prices_received"`
	RecipePricesReceived UnixTimestamp `json:"recipe_prices_received"`
	StatsReceived        UnixTimestamp `json:"stats_received"`
}

func (timestamps ConnectedRealmTimestamps) ToList() UnixTimestamps {
	return UnixTimestamps{
		timestamps.Downloaded,
		timestamps.LiveAuctionsReceived,
		timestamps.ItemPricesReceived,
		timestamps.RecipePricesReceived,
		timestamps.StatsReceived,
	}
}

func (timestamps ConnectedRealmTimestamps) IsZero() bool {
	return timestamps.ToList().IsZero()
}

func (timestamps ConnectedRealmTimestamps) Merge(
	in ConnectedRealmTimestamps,
) ConnectedRealmTimestamps {
	if !in.Downloaded.IsZero() {
		timestamps.Downloaded = in.Downloaded
	}

	if !in.LiveAuctionsReceived.IsZero() {
		timestamps.LiveAuctionsReceived = in.LiveAuctionsReceived
	}

	if !in.ItemPricesReceived.IsZero() {
		timestamps.ItemPricesReceived = in.ItemPricesReceived
	}

	if !in.RecipePricesReceived.IsZero() {
		timestamps.RecipePricesReceived = in.RecipePricesReceived
	}

	if !in.StatsReceived.IsZero() {
		timestamps.StatsReceived = in.StatsReceived
	}

	return timestamps
}
