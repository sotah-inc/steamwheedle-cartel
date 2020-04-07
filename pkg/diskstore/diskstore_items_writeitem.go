package diskstore

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func (ds DiskStore) WriteItem(id blizzardv2.ItemId, data []byte) error {
	return util.WriteFile(ds.resolveItemFilepath(id), data)
}

type WriteItemsOutJob struct {
	Err  error
	Item sotah.Item
}

func (ds DiskStore) WriteItems(in chan sotah.Item) chan WriteItemsOutJob {
	// establishing channels
	out := make(chan WriteItemsOutJob)

	// spinning up the workers for writing
	worker := func() {
		for item := range in {
			encodedItem, err := item.EncodeForStorage()
			if err != nil {
				out <- WriteItemsOutJob{
					Err:  err,
					Item: item,
				}

				continue
			}

			if err := ds.WriteItem(item.BlizzardMeta.Id, encodedItem); err != nil {
				out <- WriteItemsOutJob{
					Err:  err,
					Item: item,
				}

				continue
			}

			out <- WriteItemsOutJob{
				Err:  nil,
				Item: item,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	// queueing up the jobs
	go func() {
		for item := range in {
			logging.WithFields(logrus.Fields{
				"item": item.BlizzardMeta.Id,
			}).Debug("queueing up item for writing")

			in <- item
		}
	}()

	return out
}
