package state

import (
	"encoding/base64"
	"encoding/json"

	nats "github.com/nats-io/nats.go"
	TokensDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/tokens" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type ShortTokenHistoryResponse struct {
	History TokensDatabase.ShortTokenHistory `json:"history"`
}

func (res ShortTokenHistoryResponse) EncodeForDelivery() (string, error) {
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

func (sta TokensState) ListenForTokenHistory(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.TokenHistory), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		history, err := sta.TokensDatabase.GetShortTokenHistory(sta.Regions.Names())
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		res := ShortTokenHistoryResponse{History: history}
		data, err := res.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = data

		// dumping it out
		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
