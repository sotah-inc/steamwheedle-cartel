package act

import (
	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
)

type ComputeLiveAuctionsInJob struct {
	sotah.RegionRealmTimestampTuple
}

type ComputeLiveAuctionsOutJob struct {
	sotah.RegionRealmTuple
	Data ResponseMeta
	Err  error
}

func (job ComputeLiveAuctionsOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":  job.Err.Error(),
		"region": job.RegionName,
		"realm":  job.RealmSlug,
	}
}

func (c Client) ComputeLiveAuctions(tuples sotah.RegionRealmTimestampTuples) chan ComputeLiveAuctionsOutJob {
	// establishing channels
	in := make(chan ComputeLiveAuctionsInJob)
	out := make(chan ComputeLiveAuctionsOutJob)

	// spinning up the workers
	worker := func() {
		for inJob := range in {
			body, err := inJob.RegionRealmTimestampTuple.EncodeForDelivery()
			if err != nil {
				out <- ComputeLiveAuctionsOutJob{
					RegionRealmTuple: inJob.RegionRealmTuple,
					Data:             ResponseMeta{},
					Err:              err,
				}

				continue
			}

			actData, err := c.Call("/", "POST", []byte(body))
			if err != nil {
				out <- ComputeLiveAuctionsOutJob{
					RegionRealmTuple: inJob.RegionRealmTuple,
					Data:             ResponseMeta{},
					Err:              err,
				}

				continue
			}

			out <- ComputeLiveAuctionsOutJob{
				RegionRealmTuple: inJob.RegionRealmTuple,
				Data:             actData,
				Err:              nil,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	// queueing up the regions
	go func() {
		for _, tuple := range tuples {
			in <- ComputeLiveAuctionsInJob{
				RegionRealmTimestampTuple: tuple,
			}
		}

		close(in)
	}()

	return out
}
