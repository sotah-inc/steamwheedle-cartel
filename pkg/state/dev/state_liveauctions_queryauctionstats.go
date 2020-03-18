package dev

import (
	"fmt"

	"github.com/nats-io/go-nats"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (laState LiveAuctionsState) ListenForQueryAuctionStats(stop state.ListenStopChan) error {
	err := laState.IO.Messenger.Subscribe(string(subjects.QueryAuctionStats), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		req, err := state.NewAuctionsStatsRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// fetching aggregated stats across all regions and realms
		if req.RegionName == "" {
			totalStats := database.AuctionStats{}

			for _, status := range laState.Statuses {
				for job := range laState.IO.Databases.LiveAuctionsDatabases.GetAuctionStats(status.Realms) {
					if job.Err != nil {
						m.Err = job.Err.Error()
						m.Code = mCodes.GenericError
						laState.IO.Messenger.ReplyTo(natsMsg, m)

						return
					}

					totalStats = totalStats.Append(job.AuctionStats)
				}
			}

			laState.sendQueryAuctionStatsResponse(natsMsg, m, totalStats)

			return
		}

		statuses, ok := laState.Statuses[blizzard.RegionName(req.RegionName)]
		if !ok {
			m.Err = fmt.Sprintf("Region %s not found in statuses", req.RegionName)
			m.Code = mCodes.NotFound
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// fetching aggregated status across one region
		if req.RealmSlug == "" {
			totalStats := database.AuctionStats{}

			for job := range laState.IO.Databases.LiveAuctionsDatabases.GetAuctionStats(statuses.Realms) {
				if job.Err != nil {
					m.Err = job.Err.Error()
					m.Code = mCodes.GenericError
					laState.IO.Messenger.ReplyTo(natsMsg, m)

					return
				}

				totalStats = totalStats.Append(job.AuctionStats)
			}

			laState.sendQueryAuctionStatsResponse(natsMsg, m, totalStats)

			return
		}

		// fetching stats for one realm
		realmDb, err := laState.IO.Databases.LiveAuctionsDatabases.GetDatabase(
			blizzard.RegionName(req.RegionName),
			blizzard.RealmSlug(req.RealmSlug),
		)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		auctionStats, err := realmDb.GetAuctionStats()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		laState.sendQueryAuctionStatsResponse(natsMsg, m, auctionStats)
	})
	if err != nil {
		return err
	}

	return nil
}

func (laState LiveAuctionsState) sendQueryAuctionStatsResponse(
	natsMsg nats.Msg,
	m messenger.Message,
	auctionStats database.AuctionStats,
) {
	encoded, err := auctionStats.EncodeForDelivery()
	if err != nil {
		m.Err = err.Error()
		m.Code = mCodes.GenericError
		laState.IO.Messenger.ReplyTo(natsMsg, m)

		return
	}

	m.Data = string(encoded)
	laState.IO.Messenger.ReplyTo(natsMsg, m)
}
