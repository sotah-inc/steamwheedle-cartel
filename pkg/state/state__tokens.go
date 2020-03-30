package state

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"

	nats "github.com/nats-io/go-nats"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/metric"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type TokensState struct {
	BlizzardState

	messenger      messenger.Messenger
	tokensDatabase database.TokensDatabase
	reporter       metric.Reporter
}

func (sta TokensState) ListenForTokenHistory(stop ListenStopChan) error {
	err := sta.messenger.Subscribe(string(subjects.TokenHistory), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		request, err := NewTokenHistoryRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.MsgJSONParseError
			sta.messenger.ReplyTo(natsMsg, m)

			return
		}

		// fetching token-history with request data
		tHistory, err := sta.tokensDatabase.GetHistory(blizzardv2.RegionName(request.RegionName))
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.MsgJSONParseError
			sta.messenger.ReplyTo(natsMsg, m)

			return
		}

		// marshalling for messenger
		encodedMessage, err := tHistory.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.messenger.ReplyTo(natsMsg, m)

			return
		}

		// dumping it out
		m.Data = string(encodedMessage)
		sta.messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}

func (sta TokensState) CollectRegionTokens(regions sotah.RegionList) {
	logging.Info("Collecting region-tokens")

	// going over the list of regions
	startTime := time.Now()

	// gathering tokens
	tokens, err := sta.ResolveTokens(regions)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to fetch tokens")

		return
	}

	// formatting appropriately
	regionTokenHistory := database.RegionTokenHistory{}
	for regionName, token := range tokens {
		regionTokenHistory[regionName] = database.TokenHistory{token.LastUpdatedTimestamp: token.Price}
	}

	// persisting
	if err := sta.tokensDatabase.PersistHistory(regionTokenHistory); err != nil {
		logging.WithField("error", err.Error()).Error("failed to persist region token-histories")

		return
	}

	duration := time.Since(startTime)
	sta.reporter.Report(metric.Metrics{
		"tokenscollector_intake_duration": int(duration) / 1000 / 1000 / 1000,
	})
	logging.Info("finished tokens-collector")
}
