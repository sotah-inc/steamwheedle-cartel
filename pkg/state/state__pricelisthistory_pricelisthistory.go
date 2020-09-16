package state

import (
	"encoding/base64"
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"

	"github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewPricelistHistoryRequest(data []byte) (PricelistHistoryRequest, error) {
	req := &PricelistHistoryRequest{}
	err := json.Unmarshal(data, &req)
	if err != nil {
		return PricelistHistoryRequest{}, err
	}

	return *req, nil
}

type PricelistHistoryRequest struct {
	Tuple       blizzardv2.RegionConnectedRealmTuple `json:"tuple"`
	ItemIds     blizzardv2.ItemIds                   `json:"item_ids"`
	LowerBounds sotah.UnixTimestamp                  `json:"lower_bounds"`
	UpperBounds sotah.UnixTimestamp                  `json:"upper_bounds"`
}

type PricelistHistoryResponse struct {
	History sotah.ItemPriceHistories `json:"history"`
}

func (res PricelistHistoryResponse) EncodeForDelivery() (string, error) {
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

func (sta PricelistHistoryState) ListenForPriceListHistory(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.PriceListHistory), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		req, err := NewPricelistHistoryRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		logging.WithField("req", req).Info("received request")

		shards, err := sta.PricelistHistoryDatabases.GetShards(req.Tuple)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.UserError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		logging.WithField("shards", len(shards)).Info("found shards")

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

		res := PricelistHistoryResponse{History: itemPriceHistories}

		data, err := res.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = string(data)

		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
