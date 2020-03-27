package blizzardv2

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type ItemMediaResponses []ItemMediaResponse

type GetItemMediasOutJob struct {
	Err               error
	Id                ItemId
	ItemMediaResponse ItemMediaResponse
}

func (job GetItemMediasOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"id":    job.Id,
	}
}

func (response ItemMediaResponses) GetItemMedias(
	regionHostname string,
	regionName RegionName,
	getItemURL GetItemURLFunc,
) chan GetItemMediasOutJob {
	// starting up workers for gathering individual items
	in := make(chan ItemId)
	out := make(chan GetItemMediasOutJob)
	worker := func() {
		for id := range in {
			getItemUri, err := getItemURL(regionHostname, id, regionName)
			if err != nil {
				out <- GetItemMediasOutJob{
					Err:               err,
					Id:                id,
					ItemMediaResponse: ItemMediaResponse{},
				}

				continue
			}

			itemMediaResponse, _, err := NewItemMediaFromHTTP(getItemUri)
			if err != nil {
				out <- GetItemMediasOutJob{
					Err:               err,
					Id:                id,
					ItemMediaResponse: ItemMediaResponse{},
				}

				continue
			}

			out <- GetItemMediasOutJob{
				Err:               nil,
				Id:                id,
				ItemMediaResponse: itemMediaResponse,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	return out
}
