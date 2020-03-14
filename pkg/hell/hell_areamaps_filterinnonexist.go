package hell

import (
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type FilterInNonExistJob struct {
	Id  sotah.AreaMapId
	Err error
}

func (job FilterInNonExistJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"id":  job.Id,
		"err": job.Err.Error(),
	}
}

func (c Client) FilterInNonExist(
	gameVersion gameversions.GameVersion,
	ids []sotah.AreaMapId,
) ([]sotah.AreaMapId, error) {
	// spawning workers
	in := make(chan sotah.AreaMapId)
	out := make(chan FilterInNonExistJob)
	worker := func() {
		for id := range in {
			docRef := c.Doc(getAreaMapDocumentName(gameVersion, id))
			_, err := docRef.Get(c.Context)
			if err == nil {
				continue
			}

			if status.Code(err) == codes.NotFound {
				out <- FilterInNonExistJob{
					Id:  id,
					Err: nil,
				}

				continue
			}

			out <- FilterInNonExistJob{
				Id:  id,
				Err: err,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	// spinning it up
	go func() {
		for _, id := range ids {
			in <- id
		}

		close(in)
	}()

	// waiting for results to drain out
	results := make([]sotah.AreaMapId, len(ids))
	i := 0
	for job := range out {
		if job.Err != nil {
			return []sotah.AreaMapId{}, job.Err
		}

		results[i] = job.Id

		i += 1
	}

	return results, nil
}
