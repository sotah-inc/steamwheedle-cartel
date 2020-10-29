package state

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	LiveAuctionsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/liveauctions"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type NewLiveAuctionsStateOptions struct {
	Messenger  messenger.Messenger
	LakeClient BaseLake.Client

	LiveAuctionsDatabasesDir string
	Tuples                   blizzardv2.RegionConnectedRealmTuples
	ReceiveRegionTimestamps  func(timestamps sotah.RegionTimestamps)
}

func NewLiveAuctionsState(opts NewLiveAuctionsStateOptions) (LiveAuctionsState, error) {
	dirList := []string{
		opts.LiveAuctionsDatabasesDir,
		fmt.Sprintf("%s/live-auctions", opts.LiveAuctionsDatabasesDir),
	}
	for _, tuple := range opts.Tuples {
		dirList = append(
			dirList,
			fmt.Sprintf("%s/live-auctions/%s", opts.LiveAuctionsDatabasesDir, tuple.RegionName),
		)
	}

	// ensuring related dirs exist
	if err := util.EnsureDirsExist(dirList); err != nil {
		logging.WithField("error", err.Error()).Error("failed to ensure live-auctions database dirs exists")

		return LiveAuctionsState{}, err
	}

	ladBases, err := LiveAuctionsDatabase.NewDatabases(opts.LiveAuctionsDatabasesDir, opts.Tuples)
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
	LiveAuctionsDatabases LiveAuctionsDatabase.Databases

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
