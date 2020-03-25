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

func (response ItemResponses) GetItems(regionHostname string, getItemURL GetItemURLFunc) chan GetItemsOutJob {
	// starting up workers for gathering individual items
	in := make(chan ItemId)
	out := make(chan GetItemsOutJob)
	worker := func() {
		for id := range in {
			getItemUri, err := getItemURL(regionHostname, id)
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

	return out
}
