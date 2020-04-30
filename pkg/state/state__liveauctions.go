package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type NewLiveAuctionsStateOptions struct {
	Messenger  messenger.Messenger
	LakeClient BaseLake.Client

	LiveAuctionsDatabasesDir string
	Tuples                   blizzardv2.RegionConnectedRealmTuples
	ReceiveRegionTimestamps  func(timestamps sotah.RegionTimestamps)
}

func NewLiveAuctionsState(opts NewLiveAuctionsStateOptions) (LiveAuctionsState, error) {
	ladBases, err := database.NewLiveAuctionsDatabases(opts.LiveAuctionsDatabasesDir, opts.Tuples)
	if err != nil {
		return LiveAuctionsState{}, err
	}

	return LiveAuctionsState{
		LiveAuctionsDatabases:   ladBases,
		Messenger:               opts.Messenger,
		LakeClient:              opts.LakeClient,
		Tuples:                  opts.Tuples,
		ReceiveRegionTimestamps: opts.ReceiveRegionTimestamps,
	}, nil
}

type LiveAuctionsState struct {
	LiveAuctionsDatabases database.LiveAuctionsDatabases

	Messenger               messenger.Messenger
	LakeClient              BaseLake.Client
	Tuples                  blizzardv2.RegionConnectedRealmTuples
	ReceiveRegionTimestamps func(timestamps sotah.RegionTimestamps)
}

func (sta LiveAuctionsState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.LiveAuctionsIntake: sta.ListenForLiveAuctionsIntake,
		subjects.Auctions:           sta.ListenForAuctions,
		subjects.QueryAuctionStats:  sta.ListenForQueryAuctionStats,
		subjects.PriceList:          sta.ListenForPriceList,
	}
}
