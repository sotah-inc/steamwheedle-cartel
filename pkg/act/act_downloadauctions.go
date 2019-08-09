package act

import (
	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
)

type DownloadAuctionsInJob struct {
	sotah.RegionRealmTuple
}

type DownloadAuctionsOutJob struct {
	sotah.RegionRealmTuple
	Data ResponseMeta
	Err  error
}

func (job DownloadAuctionsOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":  job.Err.Error(),
		"region": job.RegionName,
		"realm":  job.RealmSlug,
	}
}

func (c Client) DownloadAuctions(regionRealms sotah.RegionRealms) chan DownloadAuctionsOutJob {
	// establishing channels
	in := make(chan DownloadAuctionsInJob)
	out := make(chan DownloadAuctionsOutJob)

	// spinning up the workers
	worker := func() {
		for inJob := range in {
			body, err := inJob.RegionRealmTuple.EncodeForDelivery()
			if err != nil {
				out <- DownloadAuctionsOutJob{
					RegionRealmTuple: inJob.RegionRealmTuple,
					Data:             ResponseMeta{},
					Err:              err,
				}

				continue
			}

			actData, err := c.Call("/", "POST", []byte(body))
			if err != nil {
				out <- DownloadAuctionsOutJob{
					RegionRealmTuple: inJob.RegionRealmTuple,
					Data:             ResponseMeta{},
					Err:              err,
				}

				continue
			}

			out <- DownloadAuctionsOutJob{
				RegionRealmTuple: inJob.RegionRealmTuple,
				Data:             actData,
				Err:              nil,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	// queueing up the regions
	go func() {
		for regionName, realms := range regionRealms {
			for _, realm := range realms {
				in <- DownloadAuctionsInJob{
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
