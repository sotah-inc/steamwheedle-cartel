package state

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	TokensDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/tokens" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta TokensState) ListenForTokenHistoryIntake(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.TokenHistoryIntake), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		if err := sta.tokenHistoryIntake(); err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}

func (sta TokensState) tokenHistoryIntake() error {
	logging.Info("collecting region-tokens")

	startTime := time.Now()

	// gathering tokens
	tokens, err := sta.BlizzardState.ResolveTokens(sta.Regions)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to fetch tokens")

		return err
	}

	// formatting appropriately
	regionTokenHistory := TokensDatabase.RegionTokenHistory{}
	for regionName, token := range tokens {
		regionTokenHistory[regionName] = TokensDatabase.TokenHistory{
			sotah.UnixTimestamp(token.LastUpdatedTimestamp): token.Price,
		}
	}

	if err := sta.TokensDatabase.PersistHistory(regionTokenHistory); err != nil {
		logging.WithField("error", err.Error()).Error("failed to persist tokens-history")

		return err
	}
	logging.WithFields(logrus.Fields{
		"total":          len(tokens),
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total persisted in collect-region-tokens")

	return nil
}
