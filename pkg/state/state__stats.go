package state

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	StatsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/stats"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type NewStatsStateOptions struct {
	Messenger  messenger.Messenger
	LakeClient BaseLake.Client

	StatsDatabasesDir       string
	Tuples                  blizzardv2.RegionConnectedRealmTuples
	ReceiveRegionTimestamps func(timestamps sotah.RegionTimestamps)
}

func NewStatsState(opts NewStatsStateOptions) (StatsState, error) {
	dirList := []string{
		opts.StatsDatabasesDir,
		fmt.Sprintf("%s/stats", opts.StatsDatabasesDir),
	}
	for _, tuple := range opts.Tuples {
		dirList = append(
			dirList,
			fmt.Sprintf("%s/stats/%s", opts.StatsDatabasesDir, tuple.RegionName),
		)
	}

	// ensuring related dirs exist
	if err := util.EnsureDirsExist(dirList); err != nil {
		logging.WithField("error", err.Error()).Error("failed to ensure stats database dirs exists")

		return StatsState{}, err
	}

	tBases, err := StatsDatabase.NewTupleDatabases(opts.StatsDatabasesDir, opts.Tuples)
	if err != nil {
		return StatsState{}, err
	}

	return StatsState{
		StatsTupleDatabases:     tBases,
		Messenger:               opts.Messenger,
		LakeClient:              opts.LakeClient,
		Tuples:                  opts.Tuples,
		ReceiveRegionTimestamps: opts.ReceiveRegionTimestamps,
	}, nil
}

type StatsState struct {
	StatsTupleDatabases StatsDatabase.TupleDatabases

	Messenger               messenger.Messenger
	LakeClient              BaseLake.Client
	Tuples                  blizzardv2.RegionConnectedRealmTuples
	ReceiveRegionTimestamps func(timestamps sotah.RegionTimestamps)
}

func (sta StatsState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.StatsIntake:       sta.ListenForStatsIntake,
		subjects.QueryAuctionStats: sta.ListenForQueryAuctionStats,
	}
}
