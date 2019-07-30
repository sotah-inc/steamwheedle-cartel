package run

import (
	"errors"
	"net/http"
	"time"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
)

func (sta SyncItemsState) PublishToReceiveSyncedItems(results sotah.ItemIdNameMap) sotah.Message {
	data, err := results.EncodeForDelivery()
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	replyMessage, err := sta.IO.BusClient.Request(sta.receiveSyncedItemsTopic, data, 10*time.Second)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	if replyMessage.Code != codes.Ok {
		return sotah.NewErrorMessage(errors.New("reply code was not ok"))
	}

	return sotah.NewMessage()
}

func (sta SyncItemsState) SyncExistingItem(id blizzard.ItemID) (string, error) {
	itemObj, err := sta.itemsBase.GetFirmObject(id, sta.itemsBucket)
	if err != nil {
		return "", err
	}

	item, err := sta.itemsBase.NewItem(itemObj)
	if err != nil {
		return "", err
	}

	if item.NormalizedName != "" {
		return item.NormalizedName, nil
	}

	normalizedName, err := sotah.NormalizeName(item.Name)
	if err != nil {
		return "", err
	}
	item.NormalizedName = normalizedName

	if err := sta.itemsBase.WriteItem(itemObj, item); err != nil {
		return "", err
	}

	return normalizedName, nil
}

func (sta SyncItemsState) SyncItem(id blizzard.ItemID) (string, error) {
	itemObj := sta.itemsBase.GetObject(id, sta.itemsBucket)

	exists, err := sta.itemsBase.ObjectExists(itemObj)
	if err != nil {
		return "", err
	}
	if exists {
		logging.WithField("id", id).Info("Item already exists, calling func for existing item")

		return sta.SyncExistingItem(id)
	}

	logging.WithField("id", id).Info("Downloading")
	uri, err := sta.blizzardClient.AppendAccessToken(blizzard.DefaultGetItemURL(sta.primaryRegion.Hostname, id))
	if err != nil {
		return "", err
	}

	respMeta, err := blizzard.Download(uri)
	if err != nil {
		return "", err
	}
	if respMeta.Status != http.StatusOK {
		return "", errors.New("status was not OK")
	}

	logging.WithField("id", id).Info("Parsing and encoding")
	blizzardItem, err := blizzard.NewItem(respMeta.Body)
	if err != nil {
		return "", err
	}
	item := sotah.Item{Item: blizzardItem}

	normalizedName, err := sotah.NormalizeName(item.Name)
	if err != nil {
		return "", err
	}
	item.NormalizedName = normalizedName

	// writing it out to the gcloud object
	logging.WithField("id", id).Info("Writing to items-base")

	if err := sta.itemsBase.WriteItem(itemObj, item); err != nil {
		return "", err
	}

	return normalizedName, nil
}

type SyncItemsHandleIdsJob struct {
	Err            error
	Id             blizzard.ItemID
	NormalizedName string
}

func (sta SyncItemsState) Handle(ids blizzard.ItemIds) (sotah.ItemIdNameMap, error) {
	// spawning workers
	in := make(chan blizzard.ItemID)
	out := make(chan SyncItemsHandleIdsJob)
	worker := func() {
		for id := range in {
			normalizedName, err := sta.SyncItem(id)
			if err != nil {
				out <- SyncItemsHandleIdsJob{
					Err: err,
					Id:  id,
				}

				continue
			}

			out <- SyncItemsHandleIdsJob{
				Err:            nil,
				Id:             id,
				NormalizedName: normalizedName,
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

	// waiting for the results to drain out
	results := sotah.ItemIdNameMap{}
	for outJob := range out {
		if outJob.Err != nil {
			logging.WithField("error", outJob.Err.Error()).Error("Failed to sync item")

			continue
		}

		results[outJob.Id] = outJob.NormalizedName
	}

	return results, nil
}

func (sta SyncItemsState) Run(ids blizzard.ItemIds) sotah.Message {
	results, err := sta.Handle(ids)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	return sta.PublishToReceiveSyncedItems(results)
}
