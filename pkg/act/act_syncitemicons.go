package act

import (
	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
)

type SyncItemIconsInJob struct {
	payloads sotah.IconItemsPayloads
}

type SyncItemIconsOutJob struct {
	Data ResponseMeta
	Err  error
}

func (job SyncItemIconsOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
	}
}

func (c Client) SyncItemIcons(payloadsBatches sotah.IconItemsPayloadsBatches) chan SyncItemIconsOutJob {
	// establishing channels
	in := make(chan SyncItemIconsInJob)
	out := make(chan SyncItemIconsOutJob)

	// spinning up the workers
	worker := func() {
		for inJob := range in {
			body, err := inJob.payloads.EncodeForDelivery()
			if err != nil {
				out <- SyncItemIconsOutJob{
					Data: ResponseMeta{},
					Err:  err,
				}

				continue
			}

			actData, err := c.Call("/", "POST", []byte(body))
			if err != nil {
				out <- SyncItemIconsOutJob{
					Data: ResponseMeta{},
					Err:  err,
				}

				continue
			}

			out <- SyncItemIconsOutJob{
				Data: actData,
				Err:  nil,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	// queueing up the regions
	go func() {
		for _, payloads := range payloadsBatches {
			in <- SyncItemIconsInJob{
				payloads: payloads,
			}
		}

		close(in)
	}()

	return out
}
