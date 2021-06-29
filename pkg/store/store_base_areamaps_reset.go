package store

import (
	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type ResetAreaMapsInJob struct {
	AreaId sotah.AreaMapId
}

type ResetAreaMapsOutJob struct {
	AreaId sotah.AreaMapId
	Err    error
}

func (job ResetAreaMapsOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":   job.Err.Error(),
		"area-id": job.AreaId,
	}
}

func (b AreaMapsBase) ResetAreaMaps(
	in chan ResetAreaMapsInJob,
	bkt *storage.BucketHandle,
) chan ResetAreaMapsOutJob {
	// establishing channels
	out := make(chan ResetAreaMapsOutJob)

	// spinning up workers for receiving area-map bytes and persisting it
	worker := func() {
		for job := range in {
			obj := b.GetObject(job.AreaId, bkt)

			// setting acl of area-map object to public
			acl := obj.ACL()
			if err := acl.Set(b.client.Context, storage.AllUsers, storage.RoleReader); err != nil {
				out <- ResetAreaMapsOutJob{
					AreaId: job.AreaId,
					Err:    err,
				}

				continue
			}

			out <- ResetAreaMapsOutJob{
				AreaId: job.AreaId,
				Err:    nil,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	return out
}
