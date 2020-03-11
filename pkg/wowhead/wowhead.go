package wowhead

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

const AreaMapUrlFormat = "https://wow.zamimg.com/images/wow/maps/enus/original/%d.jpg"

func DownloadAreaMap(id int) ([]byte, error) {
	return util.Download(fmt.Sprintf(AreaMapUrlFormat, id))
}

type DownloadAreaMapsJob struct {
	Err  error
	Id   int
	Data []byte
}

func (job DownloadAreaMapsJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"id":    job.Id,
	}
}

func DownloadAreaMaps(ids []int) chan DownloadAreaMapsJob {
	// spawning workers
	in := make(chan int)
	out := make(chan DownloadAreaMapsJob)
	worker := func() {
		for id := range in {
			data, err := DownloadAreaMap(id)
			if err != nil {
				out <- DownloadAreaMapsJob{
					Err:  err,
					Id:   id,
					Data: nil,
				}

				continue
			}

			out <- DownloadAreaMapsJob{
				Err:  nil,
				Id:   id,
				Data: data,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(2, worker, postWork)

	// spinning it up
	go func() {
		for _, id := range ids {
			logging.WithField("id", id).Info("Enqueueing for wowhead download")

			in <- id
		}

		close(in)
	}()

	return out
}
