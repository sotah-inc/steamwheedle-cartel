package state

import (
	"github.com/sirupsen/logrus"
	RegionsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/regions" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type NewRegionStateOptions struct {
	BlizzardState            BlizzardState
	Regions                  sotah.RegionList
	Messenger                messenger.Messenger
	RegionRealmSlugWhitelist sotah.RegionRealmSlugWhitelist
	RegionsDatabaseDir       string
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

	for _, region := range opts.Regions {
		connectedRealmIds, err := regionsDatabase.GetConnectedRealmIds(region.Name)
		if err != nil {
			return RegionsState{}, err
		}

		connectedRealmsOut, err := opts.BlizzardState.ResolveConnectedRealms(region, connectedRealmIds)
		if err != nil {
			return RegionsState{}, err
		}

		persistConnectedRealmsIn := make(chan RegionsDatabase.PersistConnectedRealmsInJob)
		go func() {
			for connectedRealmsOutJob := range connectedRealmsOut {
				connectedRealmComposite := sotah.RealmComposite{
					ConnectedRealmResponse: connectedRealmsOutJob.ConnectedRealmResponse,
					ModificationDates: sotah.ConnectedRealmTimestamps{
						Downloaded:           0,
						LiveAuctionsReceived: 0,
						ItemPricesReceived:   0,
						RecipePricesReceived: 0,
						StatsReceived:        0,
					},
				}

				data, err := connectedRealmComposite.EncodeForStorage()
				if err != nil {
					logging.WithFields(logrus.Fields{
						"err":             err.Error(),
						"connected-realm": connectedRealmComposite.ConnectedRealmResponse.Id,
					}).Error("failed to encode connected-realm for storage")

					continue
				}

				persistConnectedRealmsIn <- RegionsDatabase.PersistConnectedRealmsInJob{
					Id:   connectedRealmsOutJob.ConnectedRealmResponse.Id,
					Data: data,
				}
			}

			close(persistConnectedRealmsIn)
		}()

		if err := regionsDatabase.PersistConnectedRealms(
			region.Name,
			persistConnectedRealmsIn,
		); err != nil {
			return RegionsState{}, err
		}
	}

	return RegionsState{
		BlizzardState:   opts.BlizzardState,
		Messenger:       opts.Messenger,
		RegionsDatabase: regionsDatabase,
	}, nil
}

type RegionsState struct {
	BlizzardState   BlizzardState
	Messenger       messenger.Messenger
	RegionsDatabase RegionsDatabase.Database
}

func (sta RegionsState) ReceiveTimestamps(timestamps sotah.RegionTimestamps) {
	logging.WithField("timestamps", timestamps).Info("received timestamps")
}

func (sta RegionsState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.Status:                          sta.ListenForStatus,
		subjects.ConnectedRealms:                 sta.ListenForConnectedRealms,
		subjects.ValidateRegionConnectedRealm:    sta.ListenForValidateRegionConnectedRealm,
		subjects.ResolveConnectedRealm:           sta.ListenForResolveConnectedRealm,
		subjects.ValidateRegionRealm:             sta.ListenForValidateRegionRealm,
		subjects.QueryRealmModificationDates:     sta.ListenForQueryRealmModificationDates,
		subjects.ConnectedRealmModificationDates: sta.ListenForConnectedRealmModificationDates,
	}
}
