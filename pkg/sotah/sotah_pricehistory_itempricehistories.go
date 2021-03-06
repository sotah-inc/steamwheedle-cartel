package sotah

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"io"
	"io/ioutil"
	"strconv"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
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
		itemId := blizzardv2.ItemId(itemIdInt)

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

type ItemPriceHistories map[blizzardv2.ItemId]PriceHistory

type EncodeForPersistenceInJob struct {
	itemId       blizzardv2.ItemId
	priceHistory PriceHistory
}

type EncodeForPersistenceOutJob struct {
	Err    error
	ItemId blizzardv2.ItemId
	Data   string
}

func (ipHistories ItemPriceHistories) EncodeForPersistence() ([]byte, error) {
	in := make(chan EncodeForPersistenceInJob)
	out := make(chan EncodeForPersistenceOutJob)

	// spinning up the workers for encoding in parallel
	worker := func() {
		for inJob := range in {
			base64Encoded, err := inJob.priceHistory.EncodeForPersistence()
			if err != nil {
				out <- EncodeForPersistenceOutJob{
					Err:    err,
					ItemId: inJob.itemId,
					Data:   "",
				}

				continue
			}

			out <- EncodeForPersistenceOutJob{
				Err:    nil,
				Data:   base64Encoded,
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
	csvData := make([][]string, len(ipHistories))
	i := 0
	for outJob := range out {
		if outJob.Err != nil {
			return []byte{}, outJob.Err
		}

		csvData[i] = []string{
			strconv.Itoa(int(outJob.ItemId)),
			outJob.Data,
		}

		i += 1
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
