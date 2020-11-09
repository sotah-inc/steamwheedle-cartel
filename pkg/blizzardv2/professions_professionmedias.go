package blizzardv2

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetProfessionMediasInJob struct {
	ProfessionResponse ProfessionResponse
}

type GetProfessionMediasOutJob struct {
	Err                     error
	ProfessionResponse      ProfessionResponse
	ProfessionMediaResponse ProfessionMediaResponse
}

func (job GetProfessionMediasOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":      job.Err.Error(),
		"profession": job.ProfessionResponse.Id,
	}
}

func GetProfessionsMedias(
	in chan GetProfessionMediasInJob,
	getProfessionMediaURL func(string) (string, error),
) chan GetProfessionMediasOutJob {
	// starting up workers for gathering individual professions
	out := make(chan GetProfessionMediasOutJob)
	worker := func() {
		for job := range in {
			professionMediaUrl, err := getProfessionMediaURL(job.ProfessionResponse.Media.Key.Href)
			if err != nil {
				out <- GetProfessionMediasOutJob{
					Err:                     err,
					ProfessionResponse:      job.ProfessionResponse,
					ProfessionMediaResponse: ProfessionMediaResponse{},
				}

				continue
			}

			professionMediaResponse, _, err := NewProfessionMediaResponseFromHTTP(professionMediaUrl)
			if err != nil {
				out <- GetProfessionMediasOutJob{
					Err:                     err,
					ProfessionResponse:      job.ProfessionResponse,
					ProfessionMediaResponse: ProfessionMediaResponse{},
				}

				continue
			}

			out <- GetProfessionMediasOutJob{
				Err:                     nil,
				ProfessionResponse:      job.ProfessionResponse,
				ProfessionMediaResponse: professionMediaResponse,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	return out
}
