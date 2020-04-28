package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type NewPricelistHistoryStateOptions struct {
	Messenger  messenger.Messenger
	LakeClient BaseLake.Client

	PricelistHistoryDatabasesDir string
	Tuples                       blizzardv2.RegionConnectedRealmTuples
	ReceiveRegionTimestamps      func(timestamps sotah.RegionTimestamps)
}

func NewPricelistHistoryState(opts NewPricelistHistoryStateOptions) (PricelistHistoryState, error) {
	phdBases, err := database.NewPricelistHistoryDatabases(opts.PricelistHistoryDatabasesDir, opts.Tuples)
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
	PricelistHistoryDatabases *database.PricelistHistoryDatabases

	Messenger               messenger.Messenger
	LakeClient              BaseLake.Client
	Tuples                  blizzardv2.RegionConnectedRealmTuples
	ReceiveRegionTimestamps func(timestamps sotah.RegionTimestamps)
}
