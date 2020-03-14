package hell

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func (c Client) WriteAreaMap(version gameversions.GameVersion, id sotah.AreaMapId, state state.State) error {
	areaMapRef := c.Doc(getAreaMapDocumentName(version, id))

	if _, err := areaMapRef.Set(c.Context, AreaMap{state}); err != nil {
		return err
	}

	return nil
}

type LoadAreaMapsInJob struct {
	Id    sotah.AreaMapId
	State state.State
}

type LoadAreaMapsOutJob struct {
	Err error
	Id  sotah.AreaMapId
}

func (job LoadAreaMapsOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":   job.Err.Error(),
		"area-id": job.Id,
	}
}

func (c Client) LoadAreaMaps(version gameversions.GameVersion, in chan LoadAreaMapsInJob) chan LoadAreaMapsOutJob {
	// establishing channels
	out := make(chan LoadAreaMapsOutJob)

	// spinning up workers for receiving area-map bytes and persisting it
	worker := func() {
		for job := range in {
			if err := c.WriteAreaMap(version, job.Id, job.State); err != nil {
				out <- LoadAreaMapsOutJob{
					Id:  job.Id,
					Err: err,
				}

				continue
			}

			out <- LoadAreaMapsOutJob{
				Id:  job.Id,
				Err: nil,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	return out
}
