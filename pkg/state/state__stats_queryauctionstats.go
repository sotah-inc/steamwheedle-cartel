package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"

	"github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta StatsState) ListenForQueryAuctionStats(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.QueryAuctionStats), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		tuple, err := blizzardv2.NewRegionConnectedRealmTuple(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// fetching aggregated stats across all tuples
		if tuple.RegionName == "" {
			totalStats, err := sta.StatsRegionDatabases.TotalStats(sta.Tuples.RegionNames())
			if err != nil {
				logging.WithField(
					"error",
					err.Error(),
				).Error("failed on StatsRegionDatabases.TotalStats()")

				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			sta.sendQueryAuctionStatsResponse(natsMsg, m, totalStats)

			return
		}

		// fetching aggregated status across one region
		if tuple.ConnectedRealmId == 0 {
			rBase, err := sta.StatsRegionDatabases.GetRegionDatabase(tuple.RegionName)
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			regionStats, err := rBase.Stats()
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			sta.sendQueryAuctionStatsResponse(natsMsg, m, regionStats)

			return
		}

		// fetching stats for one tuple
		tBase, err := sta.StatsTupleDatabases.GetTupleDatabase(tuple)
		if err != nil {
			logging.WithField("error", err.Error()).Error("failed on StatsTupleDatabases.GetDatabase()")

			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		auctionStats, err := tBase.Stats()
		if err != nil {
			logging.WithField("error", err.Error()).Error("failed on tBase.Stats()")

			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		sta.sendQueryAuctionStatsResponse(natsMsg, m, auctionStats)
	})
	if err != nil {
		return err
	}

	return nil
}

func (sta StatsState) sendQueryAuctionStatsResponse(
	natsMsg nats.Msg,
	m messenger.Message,
	auctionStats sotah.AuctionStats,
) {
	encoded, err := auctionStats.EncodeForDelivery()
	if err != nil {
		m.Err = err.Error()
		m.Code = mCodes.GenericError
		sta.Messenger.ReplyTo(natsMsg, m)

		return
	}

	m.Data = string(encoded)
	sta.Messenger.ReplyTo(natsMsg, m)
}
