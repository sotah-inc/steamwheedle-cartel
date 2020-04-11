package sotah

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"io"
	"io/ioutil"
	"strconv"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

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
