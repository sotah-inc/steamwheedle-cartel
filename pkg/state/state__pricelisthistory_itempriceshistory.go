package state

import (
	"encoding/base64"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewItemPricesHistoryRequest(data []byte) (ItemPricesHistoryRequest, error) {
	req := &ItemPricesHistoryRequest{}
	err := json.Unmarshal(data, &req)
	if err != nil {
		return ItemPricesHistoryRequest{}, err
	}

	return *req, nil
}

type ItemPricesHistoryRequest struct {
	Tuple       blizzardv2.RegionVersionConnectedRealmTuple `json:"tuple"`
	ItemIds     blizzardv2.ItemIds                          `json:"item_ids"`
	LowerBounds sotah.UnixTimestamp                         `json:"lower_bounds"`
	UpperBounds sotah.UnixTimestamp                         `json:"upper_bounds"`
}

type ItemPricesHistoryResponse struct {
	History sotah.ItemPriceHistories `json:"history"`
}

func (res ItemPricesHistoryResponse) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(res)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}

func (sta PricelistHistoryState) ListenForItemPricesHistory(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.ItemPricesHistory), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		req, err := NewItemPricesHistoryRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		shards, err := sta.PricelistHistoryDatabases.GetShards(req.Tuple)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.UserError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		itemPriceHistories, err := shards.Between(req.LowerBounds, req.UpperBounds).GetItemPriceHistories(
			req.ItemIds,
			req.LowerBounds,
			req.UpperBounds,
		)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		res := ItemPricesHistoryResponse{History: itemPriceHistories}

		data, err := res.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = data

		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
