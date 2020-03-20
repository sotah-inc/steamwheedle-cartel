package blizzard

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetAllItemClassesOptions struct {
	RegionHostname       string
	GetItemClassIndexURL GetItemClassIndexURLFunc
	GetItemClassURL      GetItemClassURLFunc
}

type ItemSubClassComposite struct {
	Name string
	Id   ItemSubClassId
}

type ItemClassComposite struct {
	Name           string
	Id             ItemClassId
	ItemSubClasses []ItemSubClassComposite
}

type GetAllItemClassesJob struct {
	Err                error
	Id                 int
	ItemClassComposite ItemClassComposite
}

func (job GetAllItemClassesJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"id":    job.Id,
	}
}

func GetAllItemClasses(opts GetAllItemClassesOptions) ([]ItemClassComposite, error) {
	// querying index
	icIndex, _, err := NewItemClassIndexFromHTTP(opts.GetItemClassIndexURL(opts.RegionHostname))
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to get item-class-index")

		return []ItemClassComposite{}, err
	}

	// starting up workers for gathering individual item-classes
	in := make(chan int)
	out := make(chan GetAllItemClassesJob)
	worker := func() {
		for id := range in {
			iClass, _, err := NewItemClassFromHTTP(opts.GetItemClassURL(opts.RegionHostname, id))
			if err != nil {
				out <- GetAllItemClassesJob{
					Err:                err,
					Id:                 id,
					ItemClassComposite: ItemClassComposite{},
				}

				continue
			}

			isClasses := make([]ItemSubClassComposite, len(iClass.ItemSubClasses))
			for i := 0; i < len(iClass.ItemSubClasses); i += 1 {
				isClasses[i] = ItemSubClassComposite{
					Name: iClass.ItemSubClasses[i].Name,
					Id:   iClass.ItemSubClasses[i].Id,
				}
			}

			out <- GetAllItemClassesJob{
				Err: nil,
				Id:  id,
				ItemClassComposite: ItemClassComposite{
					Name:           iClass.Name,
					Id:             iClass.ClassId,
					ItemSubClasses: isClasses,
				},
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	// queueing it up
	go func() {
		for _, iClass := range icIndex.ItemClasses {
			in <- int(iClass.Id)
		}

		close(in)
	}()

	// waiting for it all to drain out
	result := make([]ItemClassComposite, len(icIndex.ItemClasses))
	i := 0
	for outJob := range out {
		if outJob.Err != nil {
			return []ItemClassComposite{}, outJob.Err
		}

		result[i] = outJob.ItemClassComposite
		i += 1
	}

	return result, nil
}
