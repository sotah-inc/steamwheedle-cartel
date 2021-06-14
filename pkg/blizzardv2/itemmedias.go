package blizzardv2

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetItemMediasInJob struct {
	Item ItemResponse
}

type GetItemMediasOutJob struct {
	Err               error
	Item              ItemResponse
	ItemMediaResponse ItemMediaResponse
}

func (job GetItemMediasOutJob) ToLogrusFields(version gameversion.GameVersion) logrus.Fields {
	return logrus.Fields{
		"error":        job.Err.Error(),
		"item":         job.Item.Id,
		"game-version": version,
	}
}

func GetItemMedias(
	in chan GetItemMediasInJob,
	getItemMediaURL func(string) (string, error),
) chan GetItemMediasOutJob {
	// starting up workers for gathering item-medias
	out := make(chan GetItemMediasOutJob)
	worker := func() {
		for job := range in {
			itemMediaUrl, err := getItemMediaURL(job.Item.Media.Key.Href)
			if err != nil {
				out <- GetItemMediasOutJob{
					Err:               err,
					Item:              job.Item,
					ItemMediaResponse: ItemMediaResponse{},
				}

				continue
			}

			itemMediaResponse, _, err := NewItemMediaFromHTTP(itemMediaUrl)
			if err != nil {
				out <- GetItemMediasOutJob{
					Err:               err,
					Item:              job.Item,
					ItemMediaResponse: ItemMediaResponse{},
				}

				continue
			}

			out <- GetItemMediasOutJob{
				Err:               nil,
				Item:              job.Item,
				ItemMediaResponse: itemMediaResponse,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	return out
}
