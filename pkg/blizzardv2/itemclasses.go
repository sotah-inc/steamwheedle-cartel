package blizzardv2

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemclass"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetAllItemClassesOptions struct {
	GetItemClassIndexURL func() (string, error)
	GetItemClassURL      func(id itemclass.Id) (string, error)
}

type GetAllItemClassesJob struct {
	Err               error
	Id                itemclass.Id
	ItemClassResponse ItemClassResponse
}

func (job GetAllItemClassesJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"id":    job.Id,
	}
}

func GetAllItemClasses(opts GetAllItemClassesOptions) ([]ItemClassResponse, error) {
	// querying index
	uri, err := opts.GetItemClassIndexURL()
	if err != nil {
		return []ItemClassResponse{}, err
	}

	icIndex, _, err := NewItemClassIndexFromHTTP(uri)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to get item-class-index")

		return []ItemClassResponse{}, err
	}

	// starting up workers for gathering individual item-classes
	in := make(chan itemclass.Id)
	out := make(chan GetAllItemClassesJob)
	worker := func() {
		for id := range in {
			getClassUri, err := opts.GetItemClassURL(id)
			if err != nil {
				out <- GetAllItemClassesJob{
					Err:               err,
					Id:                id,
					ItemClassResponse: ItemClassResponse{},
				}

				continue
			}

			iClass, _, err := NewItemClassFromHTTP(getClassUri)
			if err != nil {
				out <- GetAllItemClassesJob{
					Err:               err,
					Id:                id,
					ItemClassResponse: ItemClassResponse{},
				}

				continue
			}

			out <- GetAllItemClassesJob{
				Err:               nil,
				Id:                id,
				ItemClassResponse: iClass,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(2, worker, postWork)

	// queueing it up
	go func() {
		for _, iClass := range icIndex.ItemClasses {
			in <- iClass.Id
		}

		close(in)
	}()

	// waiting for it all to drain out
	result := make([]ItemClassResponse, len(icIndex.ItemClasses))
	i := 0
	for outJob := range out {
		if outJob.Err != nil {
			return []ItemClassResponse{}, outJob.Err
		}

		result[i] = outJob.ItemClassResponse
		i += 1
	}

	return result, nil
}
