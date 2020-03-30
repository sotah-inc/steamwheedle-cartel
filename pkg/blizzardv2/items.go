package blizzardv2

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type ItemResponses []ItemResponse

type GetItemsOutJob struct {
	Err          error
	Id           ItemId
	ItemResponse ItemResponse
}

func (job GetItemsOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"id":    job.Id,
	}
}

type GetItemsOptions struct {
	GetItemURL func(id ItemId) (string, error)
	ItemIds    []ItemId
}

func (response ItemResponses) GetItems(opts GetItemsOptions) chan GetItemsOutJob {
	// starting up workers for gathering individual items
	in := make(chan ItemId)
	out := make(chan GetItemsOutJob)
	worker := func() {
		for id := range in {
			getItemUri, err := opts.GetItemURL(id)
			if err != nil {
				out <- GetItemsOutJob{
					Err:          err,
					Id:           id,
					ItemResponse: ItemResponse{},
				}

				continue
			}

			itemResponse, _, err := NewItemFromHTTP(getItemUri)
			if err != nil {
				out <- GetItemsOutJob{
					Err:          err,
					Id:           id,
					ItemResponse: ItemResponse{},
				}

				continue
			}

			out <- GetItemsOutJob{
				Err:          nil,
				Id:           id,
				ItemResponse: itemResponse,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	// queueing it up
	go func() {
		for _, id := range opts.ItemIds {
			in <- id
		}

		close(in)
	}()

	return out
}
