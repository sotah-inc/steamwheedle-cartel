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

		if req.RegionName == "" {
			totalStats := database.MiniAuctionListGeneralStats{}

			for _, status := range laState.Statuses {
				for job := range laState.IO.Databases.LiveAuctionsDatabases.GetStats(status.Realms) {
					if job.Err != nil {
						m.Err = job.Err.Error()
						m.Code = mCodes.GenericError
						laState.IO.Messenger.ReplyTo(natsMsg, m)

						return
					}

					totalStats = totalStats.Add(job.Stats.MiniAuctionListGeneralStats)
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

		if req.RealmSlug == "" {
			totalStats := database.MiniAuctionListGeneralStats{}

			for job := range laState.IO.Databases.LiveAuctionsDatabases.GetStats(statuses.Realms) {
				if job.Err != nil {
					m.Err = job.Err.Error()
					m.Code = mCodes.GenericError
					laState.IO.Messenger.ReplyTo(natsMsg, m)

					return
				}

				totalStats = totalStats.Add(job.Stats.MiniAuctionListGeneralStats)
			}

			laState.sendQueryAuctionStatsResponse(natsMsg, m, totalStats)

			return
		}

		regionShards, ok := laState.IO.Databases.LiveAuctionsDatabases[blizzard.RegionName(req.RegionName)]
		if !ok {
			m.Err = fmt.Sprintf("region %s not found in shards", req.RegionName)
			m.Code = mCodes.GenericError
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		realmDb, ok := regionShards[blizzard.RealmSlug(req.RealmSlug)]
		if !ok {
			m.Err = fmt.Sprintf("realm %s not found in %s shards", req.RealmSlug, req.RegionName)
			m.Code = mCodes.GenericError
			laState.IO.Messenger.ReplyTo(natsMsg, m)
		}

		stats, err := realmDb.Stats()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		laState.sendQueryAuctionStatsResponse(natsMsg, m, stats.MiniAuctionListGeneralStats)

		return
	})
	if err != nil {
		return err
	}

	return nil
}

func (laState LiveAuctionsState) sendQueryAuctionStatsResponse(natsMsg nats.Msg, m messenger.Message, totalStats database.MiniAuctionListGeneralStats) {
	encoded, err := totalStats.EncodeForDelivery()
	if err != nil {
		m.Err = err.Error()
		m.Code = mCodes.GenericError
		laState.IO.Messenger.ReplyTo(natsMsg, m)

		return
	}

	m.Data = string(encoded)
	laState.IO.Messenger.ReplyTo(natsMsg, m)
}
