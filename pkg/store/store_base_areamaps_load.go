package store

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"

	"cloud.google.com/go/storage"
)

type LoadAreaMapsInJob struct {
	AreaId sotah.AreaMapId
	Data   []byte
}

type LoadAreaMapsOutJob struct {
	AreaId sotah.AreaMapId
	Err    error
}

func (job LoadAreaMapsOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":   job.Err.Error(),
		"area-id": job.AreaId,
	}
}

func (b AreaMapsBase) LoadAreaMaps(
	in chan LoadAreaMapsInJob,
	bkt *storage.BucketHandle,
) chan LoadAreaMapsOutJob {
	// establishing channels
	out := make(chan LoadAreaMapsOutJob)

	// spinning up workers for receiving area-map bytes and persisting it
	worker := func() {
		for job := range in {
			if err := b.WriteObject(job.AreaId, job.Data, bkt); err != nil {
				out <- LoadAreaMapsOutJob{
					AreaId: job.AreaId,
					Err:    err,
				}

				continue
			}

			out <- LoadAreaMapsOutJob{
				AreaId: job.AreaId,
				Err:    nil,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(2, worker, postWork)

	return out
}
