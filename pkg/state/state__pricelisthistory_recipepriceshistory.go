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

func NewRecipePricesHistoryRequest(data []byte) (RecipePricesHistoryRequest, error) {
	req := &RecipePricesHistoryRequest{}
	err := json.Unmarshal(data, &req)
	if err != nil {
		return RecipePricesHistoryRequest{}, err
	}

	return *req, nil
}

type RecipePricesHistoryRequest struct {
	Tuple       blizzardv2.RegionVersionConnectedRealmTuple `json:"tuple"`
	RecipeIds   blizzardv2.RecipeIds                        `json:"recipe_ids"`
	LowerBounds sotah.UnixTimestamp                         `json:"lower_bounds"`
	UpperBounds sotah.UnixTimestamp                         `json:"upper_bounds"`
}

type RecipePricesHistoryResponse struct {
	History sotah.RecipePriceHistories `json:"history"`
}

func (res RecipePricesHistoryResponse) EncodeForDelivery() (string, error) {
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

func (sta PricelistHistoryState) ListenForRecipePricesHistory(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.RecipePricesHistory), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		req, err := NewRecipePricesHistoryRequest(natsMsg.Data)
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

		recipePriceHistories, err := shards.Between(
			req.LowerBounds,
			req.UpperBounds,
		).GetRecipePriceHistories(
			req.RecipeIds,
			req.LowerBounds,
			req.UpperBounds,
		)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		res := RecipePricesHistoryResponse{History: recipePriceHistories}

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
