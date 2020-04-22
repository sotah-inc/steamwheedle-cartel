package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/diskstore"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type NewLiveAuctionsStateOptions struct {
	Messenger messenger.Messenger
	DiskStore diskstore.DiskStore

	LiveAuctionsDatabasesDir string
	Tuples                   blizzardv2.RegionConnectedRealmTuples
}

func NewLiveAuctionsState(opts NewLiveAuctionsStateOptions) (LiveAuctionsState, error) {
	ladBases, err := database.NewLiveAuctionsDatabases(opts.LiveAuctionsDatabasesDir, opts.Tuples)
	if err != nil {
		return LiveAuctionsState{}, err
	}

	return LiveAuctionsState{
		LiveAuctionsDatabases: ladBases,
		Messenger:             opts.Messenger,
		DiskStore:             opts.DiskStore,
	}, nil
}

type LiveAuctionsState struct {
	LiveAuctionsDatabases database.LiveAuctionsDatabases

	Messenger messenger.Messenger
	DiskStore diskstore.DiskStore
}

func (sta LiveAuctionsState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.LiveAuctionsIntake: sta.ListenForLiveAuctionsIntake,
		subjects.Auctions:           sta.ListenForAuctions,
	}
}
