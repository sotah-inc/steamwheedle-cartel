package state

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	BaseDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/base" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta PricelistHistoryState) ListenForPrunePricelistHistories(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(
		string(subjects.PrunePricelistHistories),
		stop,
		func(natsMsg nats.Msg) {
			m := messenger.NewMessage()

			if err := sta.prunePricelistHistories(); err != nil {
				m.Err = err.Error()
				m.Code = codes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			sta.Messenger.ReplyTo(natsMsg, m)
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (sta PricelistHistoryState) prunePricelistHistories() error {
	startTime := time.Now()

	// pruning databases where applicable
	if err := sta.PricelistHistoryDatabases.PruneDatabases(
		sotah.UnixTimestamp(BaseDatabase.RetentionLimit().Unix()),
	); err != nil {
		logging.WithField("error", err.Error()).Error("failed to prune pricelist-history databases")

		return err
	}

	logging.WithFields(logrus.Fields{
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("finished pruning pricelist-history databases")

	return nil
}
