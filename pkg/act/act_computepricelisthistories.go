package act

import (
	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
)

type ComputePricelistHistoriesInJob struct {
	sotah.RegionRealmTimestampTuple
}

type ComputePricelistHistoriesOutJob struct {
	sotah.RegionRealmTuple
	Data ResponseMeta
	Err  error
}

func (job ComputePricelistHistoriesOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":  job.Err.Error(),
		"region": job.RegionName,
		"realm":  job.RealmSlug,
	}
}

func (c Client) ComputePricelistHistories(
	tuples sotah.RegionRealmTimestampTuples,
) chan ComputePricelistHistoriesOutJob {
	// establishing channels
	in := make(chan ComputePricelistHistoriesInJob)
	out := make(chan ComputePricelistHistoriesOutJob)

	// spinning up the workers
	worker := func() {
		for inJob := range in {
			body, err := inJob.RegionRealmTimestampTuple.EncodeForDelivery()
			if err != nil {
				out <- ComputePricelistHistoriesOutJob{
					RegionRealmTuple: inJob.RegionRealmTuple,
					Data:             ResponseMeta{},
					Err:              err,
				}

				continue
			}

			actData, err := c.Call("/compute-pricelist-histories", "POST", []byte(body))
			if err != nil {
				out <- ComputePricelistHistoriesOutJob{
					RegionRealmTuple: inJob.RegionRealmTuple,
					Data:             ResponseMeta{},
					Err:              err,
				}

				continue
			}

			out <- ComputePricelistHistoriesOutJob{
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
			in <- ComputePricelistHistoriesInJob{
				RegionRealmTimestampTuple: tuple,
			}
		}

		close(in)
	}()

	return out
}
