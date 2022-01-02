package blizzardv2

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type ItemResponses []ItemResponse

type GetItemsOutJob struct {
	Err          error
	Status       int
	Id           ItemId
	ItemResponse ItemResponse
}

func (job GetItemsOutJob) ToLogrusFields(version gameversion.GameVersion) logrus.Fields {
	return logrus.Fields{
		"error":        job.Err.Error(),
		"status":       job.Status,
		"id":           job.Id,
		"game-version": version,
	}
}

type GetItemsOptions struct {
	GetItemURL func(id ItemId) (string, error)
	ItemIds    []ItemId
	Limit      int
}

func GetItems(opts GetItemsOptions) chan GetItemsOutJob {
	// starting up workers for gathering individual items
	in := make(chan ItemId)
	out := make(chan GetItemsOutJob)
	worker := func() {
		for id := range in {
			getItemUri, err := opts.GetItemURL(id)
			if err != nil {
				out <- GetItemsOutJob{
					Err:          err,
					Status:       0,
					Id:           id,
					ItemResponse: ItemResponse{},
				}

				continue
			}

			itemResponse, resp, err := NewItemFromHTTP(getItemUri)
			if err != nil {
				out <- GetItemsOutJob{
					Err:          err,
					Status:       resp.Status,
					Id:           id,
					ItemResponse: ItemResponse{},
				}

				continue
			}

			out <- GetItemsOutJob{
				Err:          nil,
				Status:       resp.Status,
				Id:           id,
				ItemResponse: itemResponse,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(16, worker, postWork)

	// queueing it up
	go func() {
		for i, id := range opts.ItemIds {
			logging.WithField("item-id", id).Info("enqueueing item for downloading")

			in <- id

			if i > opts.Limit {
				break
			}
		}

		close(in)
	}()

	return out
}
