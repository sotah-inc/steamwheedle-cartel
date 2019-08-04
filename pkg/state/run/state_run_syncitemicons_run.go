package run

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
)

func (sta SyncItemIconsState) PublishToReceiveSyncedItems(results sotah.ItemIdNameMap) sotah.Message {
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

func (sta SyncItemIconsState) UpdateItem(id blizzard.ItemID, objectUri string, objectName string) error {
	// gathering the object
	obj, err := sta.itemsBase.GetFirmObject(id, sta.itemsBucket)
	if err != nil {
		return err
	}

	// reading the item from the object
	item, err := sta.itemsBase.NewItem(obj)
	if err != nil {
		return err
	}

	// updating the icon data with the object uri (for non-cdn usage) and the object name (for cdn usage)
	logging.WithFields(logrus.Fields{
		"id":          id,
		"object-uri":  objectUri,
		"object-name": objectName,
	}).Info("Setting item icon fields")
	item.IconURL = objectUri
	item.IconObjectName = objectName

	jsonEncoded, err := json.Marshal(item)
	if err != nil {
		return err
	}

	gzipEncodedBody, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return err
	}

	// writing it out to the gcloud object
	logging.WithField("id", id).Info("Writing to items-base")
	wc := sta.itemsBase.GetObject(id, sta.itemsBucket).NewWriter(sta.IO.StoreClient.Context)
	wc.ContentType = "application/json"
	wc.ContentEncoding = "gzip"
	if _, err := wc.Write(gzipEncodedBody); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}

func (sta SyncItemIconsState) UpdateItems(objectUri string, objectName string, ids blizzard.ItemIds) error {
	// spawning workers
	in := make(chan blizzard.ItemID)
	out := make(chan error)
	worker := func() {
		for id := range in {
			if err := sta.UpdateItem(id, objectUri, objectName); err != nil {
				out <- err

				continue
			}

			out <- nil
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
	for err := range out {
		if err != nil {
			logging.WithField("error", err.Error()).Error("Failed to update items with icon data")

			continue
		}
	}

	return nil
}

func (sta SyncItemIconsState) SyncExistingItemIcon(payload sotah.IconItemsPayload) error {
	obj, err := sta.itemIconsBase.GetFirmObject(payload.Name, sta.itemIconsBucket)
	if err != nil {
		return err
	}

	// gathering obj attrs for generating a valid uri
	objAttrs, err := obj.Attrs(sta.IO.StoreClient.Context)
	if err != nil {
		return err
	}

	objectUri := fmt.Sprintf(store.ItemIconURLFormat, sta.itemIconsBucketName, objAttrs.Name)

	return sta.UpdateItems(objectUri, objAttrs.Name, payload.Ids)
}

func (sta SyncItemIconsState) SyncItemIcon(payload sotah.IconItemsPayload) error {
	obj := sta.itemIconsBase.GetObject(payload.Name, sta.itemIconsBucket)
	exists, err := sta.itemIconsBase.ObjectExists(obj)
	if err != nil {
		return err
	}
	if exists {
		logging.WithField("icon", payload.Name).Info("Item-icon already exists, updating items")

		return sta.SyncExistingItemIcon(payload)
	}

	logging.WithField("icon", payload.Name).Info("Downloading")
	respMeta, err := blizzard.Download(blizzard.DefaultGetItemIconURL(payload.Name))
	if err != nil {
		return err
	}
	if respMeta.Status != http.StatusOK {
		return errors.New("status was not OK")
	}

	// writing it out to the gcloud object
	logging.WithField("icon", payload.Name).Info("Writing to item-icons-base")
	wc := obj.NewWriter(sta.IO.StoreClient.Context)
	wc.ContentType = "image/jpeg"
	if _, err := wc.Write(respMeta.Body); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}

	// setting acl of item-icon object to public
	acl := obj.ACL()
	if err := acl.Set(sta.IO.StoreClient.Context, storage.AllUsers, storage.RoleReader); err != nil {
		return err
	}

	// gathering obj attrs for generating a valid uri
	objAttrs, err := obj.Attrs(sta.IO.StoreClient.Context)
	if err != nil {
		return err
	}

	objectUri := fmt.Sprintf(store.ItemIconURLFormat, sta.itemIconsBucketName, objAttrs.Name)

	return sta.UpdateItems(objectUri, objAttrs.Name, payload.Ids)
}

type SyncItemIconsHandlePayloadsJob struct {
	Err      error
	IconName string
	Ids      blizzard.ItemIds
}

func (sta SyncItemIconsState) HandlePayloads(payloads sotah.IconItemsPayloads) (blizzard.ItemIds, error) {
	// spawning workers
	in := make(chan sotah.IconItemsPayload)
	out := make(chan SyncItemIconsHandlePayloadsJob)
	worker := func() {
		for payload := range in {
			if err := sta.SyncItemIcon(payload); err != nil {
				out <- SyncItemIconsHandlePayloadsJob{
					Err:      err,
					IconName: payload.Name,
				}

				continue
			}

			out <- SyncItemIconsHandlePayloadsJob{
				Err:      nil,
				IconName: payload.Name,
				Ids:      payload.Ids,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	// spinning it up
	go func() {
		for _, payload := range payloads {
			in <- payload
		}

		close(in)
	}()

	// waiting for the results to drain out
	results := blizzard.ItemIds{}
	for outJob := range out {
		if outJob.Err != nil {
			logging.WithField("error", outJob.Err.Error()).Error("Failed to sync item")

			return blizzard.ItemIds{}, outJob.Err
		}

		results = append(results, outJob.Ids...)
	}

	return results, nil
}

func (sta SyncItemIconsState) Run(payloads sotah.IconItemsPayloads) sotah.Message {
	logging.WithField("payloads", payloads).Info("Handling")

	ids, err := sta.HandlePayloads(payloads)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	idNormalizedNameMap := sotah.ItemIdNameMap{}
	for _, id := range ids {
		idNormalizedNameMap[id] = ""
	}

	return sta.PublishToReceiveSyncedItems(idNormalizedNameMap)
}
