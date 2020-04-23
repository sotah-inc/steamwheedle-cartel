package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"

	"github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta LiveAuctionsState) ListenForQueryAuctionStats(stop ListenStopChan) error {
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
			totalStats, err := sta.LiveAuctionsDatabases.AuctionStatsWithTuples(sta.RegionsState.RegionComposites.ToTuples())
			if err != nil {
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
			totalStats, err := sta.LiveAuctionsDatabases.AuctionStatsWithTuples(
				sta.RegionsState.RegionComposites.ToTuples().FilterByRegionName(tuple.RegionName),
			)
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			sta.sendQueryAuctionStatsResponse(natsMsg, m, totalStats)

			return
		}

		// fetching stats for one tuple
		ladBase, err := sta.LiveAuctionsDatabases.GetDatabase(tuple)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		auctionStats, err := ladBase.AuctionStats()
		if err != nil {
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

func (sta LiveAuctionsState) sendQueryAuctionStatsResponse(
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
