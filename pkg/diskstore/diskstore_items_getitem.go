package diskstore

import (
	"os"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func (ds DiskStore) GetItem(id blizzardv2.ItemId) (sotah.Item, error) {
	itemFilepath := ds.resolveItemFilepath(id)
	if _, err := os.Stat(itemFilepath); err != nil {
		if !os.IsNotExist(err) {
			return sotah.Item{}, err
		}

		return sotah.Item{}, nil
	}

	gzipDecoded, err := util.ReadFile(itemFilepath)
	if err != nil {
		return sotah.Item{}, err
	}

	return sotah.NewItemFromGzipped(gzipDecoded)
}

type GetItemsJob struct {
	Err  error
	Id   blizzardv2.ItemId
	Item sotah.Item
}

func (ds DiskStore) GetItems(ids []blizzardv2.ItemId) chan GetItemsJob {
	// establishing channels
	out := make(chan GetItemsJob)
	in := make(chan blizzardv2.ItemId)

	// spinning up the workers for fetching items
	worker := func() {
		for id := range in {
			item, err := ds.GetItem(id)
			if err != nil {
				out <- GetItemsJob{
					Err:  err,
					Id:   id,
					Item: sotah.Item{},
				}

				continue
			}

			out <- GetItemsJob{
				Err:  nil,
				Id:   id,
				Item: item,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	// queueing up the items
	go func() {
		for _, ID := range ids {
			in <- ID
		}

		close(in)
	}()

	return out
}
