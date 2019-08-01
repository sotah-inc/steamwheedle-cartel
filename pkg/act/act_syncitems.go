package act

import (
	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
)

type SyncItemsInJob struct {
	itemIds blizzard.ItemIds
}

type SyncItemsOutJob struct {
	Data ResponseMeta
	Err  error
}

func (job SyncItemsOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
	}
}

func (c Client) SyncItems(idBatches sotah.ItemIdBatches) chan SyncItemsOutJob {
	// establishing channels
	in := make(chan SyncItemsInJob)
	out := make(chan SyncItemsOutJob)

	// spinning up the workers
	worker := func() {
		for inJob := range in {
			body, err := inJob.itemIds.EncodeForDelivery()
			if err != nil {
				out <- SyncItemsOutJob{
					Data: ResponseMeta{},
					Err:  err,
				}

				continue
			}

			actData, err := c.Call("/", "POST", []byte(body))
			if err != nil {
				out <- SyncItemsOutJob{
					Data: ResponseMeta{},
					Err:  err,
				}

				continue
			}

			out <- SyncItemsOutJob{
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
		for _, ids := range idBatches {
			in <- SyncItemsInJob{
				itemIds: ids,
			}
		}

		close(in)
	}()

	return out
}
