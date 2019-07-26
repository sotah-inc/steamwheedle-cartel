package act

import (
	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
)

type CleanupAuctionsInJob struct {
	sotah.RegionRealmTuple
}

type CleanupAuctionsOutJob struct {
	sotah.RegionRealmTuple
	Data ResponseMeta
	Err  error
}

func (job CleanupAuctionsOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":  job.Err.Error(),
		"region": job.RegionName,
		"realm":  job.RealmSlug,
	}
}

func (c Client) CleanupAuctions(regionRealms sotah.RegionRealms) chan CleanupAuctionsOutJob {
	// establishing channels
	in := make(chan CleanupAuctionsInJob)
	out := make(chan CleanupAuctionsOutJob)

	// spinning up the workers
	worker := func() {
		for inJob := range in {
			body, err := inJob.RegionRealmTuple.EncodeForDelivery()
			if err != nil {
				out <- CleanupAuctionsOutJob{
					RegionRealmTuple: inJob.RegionRealmTuple,
					Data:             ResponseMeta{},
					Err:              err,
				}

				continue
			}

			actData, err := c.Call("/", "POST", []byte(body))
			if err != nil {
				out <- CleanupAuctionsOutJob{
					RegionRealmTuple: inJob.RegionRealmTuple,
					Data:             ResponseMeta{},
					Err:              err,
				}

				continue
			}

			out <- CleanupAuctionsOutJob{
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
				in <- CleanupAuctionsInJob{
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
