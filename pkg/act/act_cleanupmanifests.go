package act

import (
	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
)

type CleanupManifestsInJob struct {
	sotah.RegionRealmTuple
}

type CleanupManifestsOutJob struct {
	sotah.RegionRealmTuple
	Data ResponseMeta
	Err  error
}

func (job CleanupManifestsOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":  job.Err.Error(),
		"region": job.RegionName,
		"realm":  job.RealmSlug,
	}
}

func (c Client) CleanupManifests(regionRealms sotah.RegionRealms) chan CleanupManifestsOutJob {
	// establishing channels
	in := make(chan CleanupManifestsInJob)
	out := make(chan CleanupManifestsOutJob)

	// spinning up the workers
	worker := func() {
		for inJob := range in {
			body, err := inJob.RegionRealmTuple.EncodeForDelivery()
			if err != nil {
				out <- CleanupManifestsOutJob{
					RegionRealmTuple: inJob.RegionRealmTuple,
					Data:             ResponseMeta{},
					Err:              err,
				}

				continue
			}

			actData, err := c.Call("/", "POST", []byte(body))
			if err != nil {
				out <- CleanupManifestsOutJob{
					RegionRealmTuple: inJob.RegionRealmTuple,
					Data:             ResponseMeta{},
					Err:              err,
				}

				continue
			}

			out <- CleanupManifestsOutJob{
				RegionRealmTuple: inJob.RegionRealmTuple,
				Data:             actData,
				Err:              nil,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(64, worker, postWork)

	// queueing up the regions
	go func() {
		for regionName, realms := range regionRealms {
			for _, realm := range realms {
				in <- CleanupManifestsInJob{
					RegionRealmTuple: sotah.RegionRealmTuple{
						RegionName: string(regionName),
						RealmSlug:  string(realm.Slug),
					},
				}
			}
		}

		close(in)
	}()

	return out
}
