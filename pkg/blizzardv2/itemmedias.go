package blizzardv2

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetItemMediasInJob struct {
	URL string
}

type GetItemMediasOutJob struct {
	Err               error
	ItemMediaResponse ItemMediaResponse
}

func (job GetItemMediasOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
	}
}

func GetItemMedias(
	in chan GetItemMediasInJob,
) chan GetItemMediasOutJob {
	// starting up workers for gathering item-medias
	out := make(chan GetItemMediasOutJob)
	worker := func() {
		for job := range in {
			itemMediaResponse, _, err := NewItemMediaFromHTTP(job.URL)
			if err != nil {
				out <- GetItemMediasOutJob{
					Err:               err,
					ItemMediaResponse: ItemMediaResponse{},
				}

				continue
			}

			out <- GetItemMediasOutJob{
				Err:               nil,
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
