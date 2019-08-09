package act

import (
	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
)

type CleanupPricelistHistoriesInJob struct {
	sotah.RegionRealmTuple
}

type CleanupPricelistHistoriesOutJob struct {
	sotah.RegionRealmTuple
	Data ResponseMeta
	Err  error
}

func (job CleanupPricelistHistoriesOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":  job.Err.Error(),
		"region": job.RegionName,
		"realm":  job.RealmSlug,
	}
}

func (c Client) CleanupPricelistHistories(regionRealms sotah.RegionRealms) chan CleanupPricelistHistoriesOutJob {
	// establishing channels
	in := make(chan CleanupPricelistHistoriesInJob)
	out := make(chan CleanupPricelistHistoriesOutJob)

	// spinning up the workers
	worker := func() {
		for inJob := range in {
			body, err := inJob.RegionRealmTuple.EncodeForDelivery()
			if err != nil {
				out <- CleanupPricelistHistoriesOutJob{
					RegionRealmTuple: inJob.RegionRealmTuple,
					Data:             ResponseMeta{},
					Err:              err,
				}

				continue
			}

			actData, err := c.Call("/", "POST", []byte(body))
			if err != nil {
				out <- CleanupPricelistHistoriesOutJob{
					RegionRealmTuple: inJob.RegionRealmTuple,
					Data:             ResponseMeta{},
					Err:              err,
				}

				continue
			}

			out <- CleanupPricelistHistoriesOutJob{
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
				in <- CleanupPricelistHistoriesInJob{
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
