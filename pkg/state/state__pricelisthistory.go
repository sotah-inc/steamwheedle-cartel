package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	PricelistHistoryDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/pricelisthistory"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type NewPricelistHistoryStateOptions struct {
	Messenger  messenger.Messenger
	LakeClient BaseLake.Client

	PricelistHistoryDatabasesDir string
	Tuples                       blizzardv2.RegionConnectedRealmTuples
	ReceiveRegionTimestamps      func(timestamps sotah.RegionTimestamps)
}

func NewPricelistHistoryState(opts NewPricelistHistoryStateOptions) (PricelistHistoryState, error) {
	phdBases, err := PricelistHistoryDatabase.NewDatabases(opts.PricelistHistoryDatabasesDir, opts.Tuples)
	if err != nil {
		return PricelistHistoryState{}, err
	}

	return PricelistHistoryState{
		PricelistHistoryDatabases: phdBases,
		Messenger:                 opts.Messenger,
		LakeClient:                opts.LakeClient,
		Tuples:                    opts.Tuples,
		ReceiveRegionTimestamps:   opts.ReceiveRegionTimestamps,
	}, nil
}

type PricelistHistoryState struct {
	PricelistHistoryDatabases *PricelistHistoryDatabase.Databases

	Messenger               messenger.Messenger
	LakeClient              BaseLake.Client
	Tuples                  blizzardv2.RegionConnectedRealmTuples
	ReceiveRegionTimestamps func(timestamps sotah.RegionTimestamps)
}

func (sta PricelistHistoryState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.PricelistHistoryIntake: sta.ListenForPricelistHistoryIntake,
		subjects.PriceListHistory:       sta.ListenForPriceListHistory,
	}
}
