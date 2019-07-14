package sotah

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"io"
	"io/ioutil"
	"math"
	"sort"
	"strconv"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
)

// item-prices
func NewItemPrices(maList MiniAuctionList) ItemPrices {
	itemIds := maList.ItemIds()
	iPrices := map[blizzard.ItemID]Prices{}
	itemIDMap := make(map[blizzard.ItemID]struct{}, len(itemIds))
	itemBuyoutPers := make(map[blizzard.ItemID][]float64, len(itemIds))
	for _, id := range itemIds {
		iPrices[id] = Prices{}
		itemIDMap[id] = struct{}{}
		itemBuyoutPers[id] = []float64{}
	}

	for _, mAuction := range maList {
		id := mAuction.ItemID

		if _, ok := itemIDMap[id]; !ok {
			continue
		}

		p := iPrices[id]

		if mAuction.Buyout > 0 {
			auctionBuyoutPer := float64(mAuction.Buyout / mAuction.Quantity)

			itemBuyoutPers[id] = append(itemBuyoutPers[id], auctionBuyoutPer)

			if p.MinBuyoutPer == 0 || auctionBuyoutPer < p.MinBuyoutPer {
				p.MinBuyoutPer = auctionBuyoutPer
			}
			if p.MaxBuyoutPer == 0 || auctionBuyoutPer > p.MaxBuyoutPer {
				p.MaxBuyoutPer = auctionBuyoutPer
			}
		}

		p.Volume += mAuction.Quantity * int64(len(mAuction.AucList))

		iPrices[id] = p
	}

	for id, buyouts := range itemBuyoutPers {
		if len(buyouts) == 0 {
			continue
		}

		p := iPrices[id]

		// gathering total and calculating average
		total := float64(0)
		for _, buyout := range buyouts {
			total += buyout
		}
		p.AverageBuyoutPer = total / float64(len(buyouts))

		// sorting buyouts and calculating median
		buyoutsSlice := sort.Float64Slice(buyouts)
		buyoutsSlice.Sort()
		hasEvenMembers := len(buyoutsSlice)%2 == 0
		median := func() float64 {
			if hasEvenMembers {
				middle := float64(len(buyoutsSlice)) / 2
				return (buyoutsSlice[int(math.Floor(middle))] + buyoutsSlice[int(math.Ceil(middle))]) / 2
			}

			return buyoutsSlice[(len(buyoutsSlice)-1)/2]
		}()
		p.MedianBuyoutPer = median

		iPrices[id] = p
	}

	return iPrices
}

type ItemPrices map[blizzard.ItemID]Prices

func (iPrices ItemPrices) ItemIds() []blizzard.ItemID {
	out := []blizzard.ItemID{}
	for ID := range iPrices {
		out = append(out, ID)
	}

	return out
}

// prices
func NewPricesFromBytes(data []byte) (Prices, error) {
	gzipDecoded, err := util.GzipDecode(data)
	if err != nil {
		return Prices{}, err
	}

	pricesValue := Prices{}
	if err := json.Unmarshal(gzipDecoded, &pricesValue); err != nil {
		return Prices{}, err
	}

	return pricesValue, nil
}

type Prices struct {
	MinBuyoutPer     float64 `json:"min_buyout_per"`
	MaxBuyoutPer     float64 `json:"max_buyout_per"`
	AverageBuyoutPer float64 `json:"average_buyout_per"`
	MedianBuyoutPer  float64 `json:"median_buyout_per"`
	Volume           int64   `json:"volume"`
}

func (p Prices) EncodeForPersistence() ([]byte, error) {
	jsonEncoded, err := json.Marshal(p)
	if err != nil {
		return []byte{}, err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return []byte{}, err
	}

	return gzipEncoded, nil
}

// item-price-histories
func NewItemPriceHistoriesFromMinimized(reader io.Reader) (ItemPriceHistories, error) {
	out := ItemPriceHistories{}

	r := csv.NewReader(reader)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ItemPriceHistories{}, err
		}

		itemIdInt, err := strconv.Atoi(record[0])
		if err != nil {
			return ItemPriceHistories{}, err
		}
		itemId := blizzard.ItemID(itemIdInt)

		base64DecodedPriceHistory, err := base64.StdEncoding.DecodeString(record[1])
		if err != nil {
			return ItemPriceHistories{}, err
		}

		gzipDecodedPriceHistory, err := util.GzipDecode(base64DecodedPriceHistory)
		if err != nil {
			return ItemPriceHistories{}, err
		}

		var pHistory PriceHistory
		if err := json.Unmarshal(gzipDecodedPriceHistory, &pHistory); err != nil {
			return ItemPriceHistories{}, err
		}

		out[itemId] = pHistory
	}

	return out, nil
}

type ItemPriceHistories map[blizzard.ItemID]PriceHistory

type EncodeForPersistenceInJob struct {
	itemId       blizzard.ItemID
	priceHistory PriceHistory
}

type EncodeForPersistenceOutJob struct {
	Err    error
	ItemId blizzard.ItemID
	Data   string
}

func (ipHistories ItemPriceHistories) EncodeForPersistence() ([]byte, error) {
	in := make(chan EncodeForPersistenceInJob)
	out := make(chan EncodeForPersistenceOutJob)

	// spinning up the workers for encoding in parallel
	worker := func() {
		for inJob := range in {
			jsonEncodedPriceHistory, err := json.Marshal(inJob.priceHistory)
			if err != nil {
				continue
			}

			gzipEncodedPriceHistory, err := util.GzipEncode(jsonEncodedPriceHistory)
			if err != nil {
				continue
			}

			out <- EncodeForPersistenceOutJob{
				Err:    nil,
				Data:   base64.StdEncoding.EncodeToString(gzipEncodedPriceHistory),
				ItemId: inJob.itemId,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	// queueing up the price-histories for encoding
	go func() {
		for itemId, priceHistory := range ipHistories {
			in <- EncodeForPersistenceInJob{
				itemId:       itemId,
				priceHistory: priceHistory,
			}
		}

		close(in)
	}()

	// waiting for them to drain out
	csvData := [][]string{}
	for outJob := range out {
		if outJob.Err != nil {
			return []byte{}, outJob.Err
		}

		csvData = append(csvData, []string{
			strconv.Itoa(int(outJob.ItemId)),
			outJob.Data,
		})
	}

	// producing a receiver
	buf := bytes.NewBuffer([]byte{})
	w := csv.NewWriter(buf)
	if err := w.WriteAll(csvData); err != nil {
		return []byte{}, err
	}
	if err := w.Error(); err != nil {
		return []byte{}, err
	}

	// encoding the data
	data, err := ioutil.ReadAll(buf)
	if err != nil {
		return []byte{}, err
	}
	gzipEncodedData, err := util.GzipEncode(data)
	if err != nil {
		return []byte{}, err
	}

	return gzipEncodedData, nil
}

// price-history
func NewPriceHistoryFromBytes(data []byte) (PriceHistory, error) {
	gzipDecoded, err := util.GzipDecode(data)
	if err != nil {
		return PriceHistory{}, err
	}

	out := PriceHistory{}
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return PriceHistory{}, err
	}

	return out, nil
}

type PriceHistory map[UnixTimestamp]Prices

func (pHistory PriceHistory) EncodeForPersistence() ([]byte, error) {
	jsonEncoded, err := json.Marshal(pHistory)
	if err != nil {
		return []byte{}, err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return []byte{}, err
	}

	return gzipEncoded, nil
}

type PricelistHistoryVersions map[blizzard.RegionName]map[blizzard.RealmSlug]map[UnixTimestamp]string

func (v PricelistHistoryVersions) Insert(
	regionName blizzard.RegionName,
	realmSlug blizzard.RealmSlug,
	targetTimestamp UnixTimestamp,
	version string,
) PricelistHistoryVersions {
	if _, ok := v[regionName]; !ok {
		v[regionName] = map[blizzard.RealmSlug]map[UnixTimestamp]string{}
	}
	if _, ok := v[regionName][realmSlug]; !ok {
		v[regionName][realmSlug] = map[UnixTimestamp]string{}
	}

	v[regionName][realmSlug][targetTimestamp] = version

	return v
}

func NewCleanupPricelistPayloads(regionRealmMap map[blizzard.RegionName]Realms) CleanupPricelistPayloads {
	out := CleanupPricelistPayloads{}
	for regionName, realms := range regionRealmMap {
		for _, realm := range realms {
			out = append(out, CleanupPricelistPayload{
				RegionName: string(regionName),
				RealmSlug:  string(realm.Slug),
			})
		}
	}

	return out
}

type CleanupPricelistPayloads []CleanupPricelistPayload

func NewCleanupPricelistPayload(data string) (CleanupPricelistPayload, error) {
	var out CleanupPricelistPayload
	if err := json.Unmarshal([]byte(data), &out); err != nil {
		return CleanupPricelistPayload{}, err
	}

	return out, nil
}

type CleanupPricelistPayload struct {
	RegionName string `json:"region_name"`
	RealmSlug  string `json:"realm_slug"`
}

func (p CleanupPricelistPayload) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	return string(jsonEncoded), nil
}

func NewCleanupPricelistPayloadResponse(data string) (CleanupPricelistPayloadResponse, error) {
	var out CleanupPricelistPayloadResponse
	if err := json.Unmarshal([]byte(data), &out); err != nil {
		return CleanupPricelistPayloadResponse{}, err
	}

	return out, nil
}

type CleanupPricelistPayloadResponse struct {
	RegionName   string `json:"region_name"`
	RealmSlug    string `json:"realm_slug"`
	TotalDeleted int    `json:"total_removed"`
}

func (p CleanupPricelistPayloadResponse) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	return string(jsonEncoded), nil
}
