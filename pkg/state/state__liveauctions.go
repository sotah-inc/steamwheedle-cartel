package state

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	LiveAuctionsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/liveauctions" // nolint:lll
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
	Tuples                   blizzardv2.RegionVersionConnectedRealmTuples
	ReceiveRegionTimestamps  func(timestamps sotah.RegionVersionTimestamps) error
}

func NewLiveAuctionsState(opts NewLiveAuctionsStateOptions) (LiveAuctionsState, error) {
	dirList := []string{
		opts.LiveAuctionsDatabasesDir,
		fmt.Sprintf("%s/live-auctions", opts.LiveAuctionsDatabasesDir),
	}
	for _, tuple := range opts.Tuples {
		dirList = append(
			dirList,
			fmt.Sprintf(
				"%s/live-auctions/%s/%s",
				opts.LiveAuctionsDatabasesDir,
				tuple.RegionName,
				tuple.Version,
			),
		)
	}

	for _, dirName := range dirList {
		logging.WithField("dir-name", dirName).Info("dir for ensuring")
	}

	// ensuring related dirs exist
	if err := util.EnsureDirsExist(dirList); err != nil {
		logging.WithField(
			"error",
			err.Error(),
		).Error("failed to ensure live-auctions database dirs exists")

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
		ReceiveRegionTimestamps: opts.ReceiveRegionTimestamps,
	}, nil
}

type LiveAuctionsState struct {
	LiveAuctionsDatabases LiveAuctionsDatabase.Databases

	Messenger               messenger.Messenger
	LakeClient              BaseLake.Client
	ReceiveRegionTimestamps func(timestamps sotah.RegionVersionTimestamps) error
}

func (sta LiveAuctionsState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.LiveAuctionsIntake: sta.ListenForLiveAuctionsIntake,
		subjects.Auctions:           sta.ListenForAuctions,
		subjects.PriceList:          sta.ListenForPriceList,
	}
}
