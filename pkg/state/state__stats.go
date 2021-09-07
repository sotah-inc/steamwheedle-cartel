package state

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	StatsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/stats" // nolint:lll
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
	Tuples                  blizzardv2.RegionVersionConnectedRealmTuples
	ReceiveRegionTimestamps func(timestamps sotah.RegionVersionTimestamps) error
}

func NewStatsState(opts NewStatsStateOptions) (StatsState, error) {
	for _, tuple := range opts.Tuples {
		logging.WithField("tuple", tuple.String()).Info("received tuple in NewStatsState()")
	}

	dirList := []string{
		opts.StatsDatabasesDir,
		fmt.Sprintf("%s/stats", opts.StatsDatabasesDir),
	}
	for _, tuple := range opts.Tuples {
		dirList = append(
			dirList,
			StatsDatabase.TupleDatabaseDirPath(
				fmt.Sprintf("%s/stats", opts.StatsDatabasesDir),
				tuple,
			),
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

	rBases, err := StatsDatabase.NewRegionDatabases(opts.StatsDatabasesDir, opts.Tuples.RegionNames())
	if err != nil {
		return StatsState{}, err
	}

	return StatsState{
		StatsTupleDatabases:     tBases,
		StatsRegionDatabases:    rBases,
		Messenger:               opts.Messenger,
		LakeClient:              opts.LakeClient,
		ReceiveRegionTimestamps: opts.ReceiveRegionTimestamps,
	}, nil
}

type StatsState struct {
	StatsTupleDatabases  StatsDatabase.TupleDatabases
	StatsRegionDatabases StatsDatabase.RegionDatabases

	Messenger               messenger.Messenger
	LakeClient              BaseLake.Client
	Tuples                  blizzardv2.RegionVersionConnectedRealmTuples
	ReceiveRegionTimestamps func(
		timestamps sotah.RegionVersionTimestamps,
	) error
}

func (sta StatsState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.StatsIntake:       sta.ListenForStatsIntake,
		subjects.QueryAuctionStats: sta.ListenForQueryAuctionStats,
	}
}
