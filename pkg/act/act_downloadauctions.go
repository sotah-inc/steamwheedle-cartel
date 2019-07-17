package act

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
)

type DownloadAuctionsInJob struct {
	RegionName blizzard.RegionName
	RealmSlug  blizzard.RealmSlug
}

type DownloadAuctionsOutJob struct {
	RegionName blizzard.RegionName
	RealmSlug  blizzard.RealmSlug
	Data       []byte
	Err        error
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
			actData, err := c.Call(
				fmt.Sprintf("/?region=%s&realm=%s", inJob.RegionName, inJob.RealmSlug),
				"GET",
				nil,
			)
			if err != nil {
				out <- DownloadAuctionsOutJob{
					RegionName: inJob.RegionName,
					RealmSlug:  inJob.RealmSlug,
					Data:       nil,
					Err:        err,
				}

				continue
			}

			out <- DownloadAuctionsOutJob{
				RegionName: inJob.RegionName,
				RealmSlug:  inJob.RealmSlug,
				Data:       actData.Body,
				Err:        nil,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	// queueing up the regions
	go func() {
		for regionName, realms := range regionRealms {
			for _, realm := range realms {
				in <- DownloadAuctionsInJob{RegionName: regionName, RealmSlug: realm.Slug}
			}
		}

		close(in)
	}()

	return out
}
