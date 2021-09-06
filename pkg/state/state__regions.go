package state

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	RegionsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/regions" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/statuskinds"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type NewRegionStateOptions struct {
	BlizzardState      BlizzardState
	Regions            sotah.RegionList
	GameVersionList    gameversion.List
	Messenger          messenger.Messenger
	RealmSlugWhitelist sotah.RealmSlugWhitelist
	RegionsDatabaseDir string
}

func NewRegionState(opts NewRegionStateOptions) (RegionsState, error) {
	regionsDatabase, err := RegionsDatabase.NewDatabase(opts.RegionsDatabaseDir)
	if err != nil {
		return RegionsState{}, err
	}

	names, err := regionsDatabase.GetRegionNames()
	if err != nil {
		return RegionsState{}, err
	}

	regionsToPersist := opts.Regions.FilterOut(names)
	if len(regionsToPersist) > 0 {
		if err := regionsDatabase.PersistRegions(regionsToPersist); err != nil {
			return RegionsState{}, err
		}
	}

	logging.WithField("regions", opts.Regions).Info("going over regions")
	for _, region := range opts.Regions {
		logging.WithFields(logrus.Fields{
			"region":        region.Name,
			"game-versions": opts.GameVersionList,
		}).Info("going over game-versions")

		for _, version := range opts.GameVersionList {
			resolvedWhitelist := opts.RealmSlugWhitelist.Get(region.Name, version)
			connectedRealmIds, err := regionsDatabase.GetConnectedRealmIds(blizzardv2.RegionVersionTuple{
				RegionTuple: blizzardv2.RegionTuple{RegionName: region.Name},
				Version:     version,
			})
			if err != nil {
				return RegionsState{}, err
			}

			if len(connectedRealmIds) > 0 {
				logging.WithField(
					"connected-realms",
					len(connectedRealmIds),
				).Info("connected-realms already present for region, skipping")

				continue
			}

			connectedRealmsOut, err := opts.BlizzardState.ResolveConnectedRealms(
				region,
				version,
				connectedRealmIds,
				resolvedWhitelist,
			)
			if err != nil {
				return RegionsState{}, err
			}

			persistConnectedRealmsIn := make(chan RegionsDatabase.PersistConnectedRealmsInJob)
			persistConnectedRealmsErrOut := make(chan error)
			go func() {
				for connectedRealmsOutJob := range connectedRealmsOut {
					if connectedRealmsOutJob.Err != nil {
						logging.WithField(
							"error",
							connectedRealmsOutJob.Err.Error(),
						).Error("failed to resolve connected-realm")

						persistConnectedRealmsErrOut <- connectedRealmsOutJob.Err

						continue
					}

					connectedRealmComposite := sotah.RealmComposite{
						ConnectedRealmResponse: connectedRealmsOutJob.ConnectedRealmResponse,
						StatusTimestamps: sotah.StatusTimestamps{
							statuskinds.Downloaded:           0,
							statuskinds.LiveAuctionsReceived: 0,
							statuskinds.ItemPricesReceived:   0,
							statuskinds.RecipePricesReceived: 0,
							statuskinds.StatsReceived:        0,
						},
					}

					data, err := connectedRealmComposite.EncodeForStorage()
					if err != nil {
						logging.WithFields(logrus.Fields{
							"err":             err.Error(),
							"connected-realm": connectedRealmComposite.ConnectedRealmResponse.Id,
						}).Error("failed to encode connected-realm for storage")

						persistConnectedRealmsErrOut <- err

						continue
					}

					persistConnectedRealmsIn <- RegionsDatabase.PersistConnectedRealmsInJob{
						Id:   connectedRealmsOutJob.ConnectedRealmResponse.Id,
						Data: data,
					}
				}

				persistConnectedRealmsErrOut <- nil
				close(persistConnectedRealmsIn)
			}()

			logging.Info("waiting for persistConnectedRealmsErrOut")
			if err := <-persistConnectedRealmsErrOut; err != nil {
				return RegionsState{}, err
			}

			logging.Info("calling regionsDatabase.PersistConnectedRealms()")
			if err := regionsDatabase.PersistConnectedRealms(
				blizzardv2.RegionVersionTuple{
					RegionTuple: blizzardv2.RegionTuple{RegionName: region.Name},
					Version:     version,
				},
				persistConnectedRealmsIn,
			); err != nil {
				return RegionsState{}, err
			}
		}
	}

	return RegionsState{
		BlizzardState:   opts.BlizzardState,
		Messenger:       opts.Messenger,
		RegionsDatabase: regionsDatabase,
		GameVersionList: opts.GameVersionList,
	}, nil
}

type RegionsState struct {
	BlizzardState   BlizzardState
	Messenger       messenger.Messenger
	RegionsDatabase RegionsDatabase.Database

	GameVersionList gameversion.List
}

func (sta RegionsState) ReceiveTimestamps(
	regionTimestamps sotah.RegionVersionTimestamps,
) error {
	return sta.RegionsDatabase.ReceiveRegionTimestamps(regionTimestamps)
}

func (sta RegionsState) ResolveDownloadTuples() ([]blizzardv2.DownloadConnectedRealmTuple, error) {
	return sta.RegionsDatabase.GetDownloadTuples()
}

func (sta RegionsState) ResolveTuples() (blizzardv2.RegionVersionConnectedRealmTuples, error) {
	return sta.RegionsDatabase.GetTuples()
}

func (sta RegionsState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.ConnectedRealmModificationDates: sta.ListenForConnectedRealmModificationDates,
		subjects.ConnectedRealms:                 sta.ListenForConnectedRealms,
		subjects.QueryRealmModificationDates:     sta.ListenForQueryRealmModificationDates,
		subjects.ReceiveRegionTimestamps:         sta.ListenForReceiveRegionTimestamps,
		subjects.ResolveConnectedRealm:           sta.ListenForResolveConnectedRealm,
		subjects.Status:                          sta.ListenForStatus,
		subjects.ValidateGameVersion:             sta.ListenForValidateGameVersion,
		subjects.ValidateRegion:                  sta.ListenForValidateRegion,
		subjects.ValidateRegionConnectedRealm:    sta.ListenForValidateRegionConnectedRealm,
		subjects.ValidateRegionRealm:             sta.ListenForValidateRegionRealm,
	}
}
