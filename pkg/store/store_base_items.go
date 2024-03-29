package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/store/regions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewItemsBase(c Client, location regions.Region, version gameversions.GameVersion) ItemsBase {
	return ItemsBase{
		base{client: c, location: location},
		version,
	}
}

type ItemsBase struct {
	base
	GameVersion gameversions.GameVersion
}

func (b ItemsBase) getBucketName() string {
	return "sotah-items"
}

func (b ItemsBase) GetBucket() *storage.BucketHandle {
	return b.base.getBucket(b.getBucketName())
}

func (b ItemsBase) GetFirmBucket() (*storage.BucketHandle, error) {
	return b.base.getFirmBucket(b.getBucketName())
}

func (b ItemsBase) getObjectName(id blizzardv2.ItemId) string {
	return fmt.Sprintf("%s/%d.json.gz", b.GameVersion, id)
}

func (b ItemsBase) GetObject(
	id blizzardv2.ItemId,
	bkt *storage.BucketHandle,
) *storage.ObjectHandle {
	return b.base.getObject(b.getObjectName(id), bkt)
}

func (b ItemsBase) GetFirmObject(
	id blizzardv2.ItemId,
	bkt *storage.BucketHandle,
) (*storage.ObjectHandle, error) {
	return b.base.getFirmObject(b.getObjectName(id), bkt)
}

func (b ItemsBase) NewItem(obj *storage.ObjectHandle) (sotah.Item, error) {
	reader, err := obj.NewReader(b.client.Context)
	if err != nil {
		return sotah.Item{}, err
	}
	defer func() {
		if err := reader.Close(); err != nil {
			logging.WithField("error", err.Error()).Error("Failed to close reader")
		}
	}()

	body, err := ioutil.ReadAll(reader) // automatically gzip-decoded
	if err != nil {
		return sotah.Item{}, err
	}

	return sotah.NewItem(body)
}

type GetItemsOutJob struct {
	Err             error
	Id              blizzardv2.ItemId
	GzipEncodedData []byte
}

func (job GetItemsOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"id":    job.Id,
	}
}

func (b ItemsBase) GetItems(
	ids []blizzardv2.ItemId,
	bkt *storage.BucketHandle,
) chan GetItemsOutJob {
	// spinning up workers
	in := make(chan blizzardv2.ItemId)
	out := make(chan GetItemsOutJob)
	worker := func() {
		for id := range in {
			obj, err := b.GetFirmObject(id, bkt)
			if err != nil {
				out <- GetItemsOutJob{
					Err: err,
					Id:  id,
				}

				continue
			}

			reader, err := obj.ReadCompressed(true).NewReader(b.client.Context)
			if err != nil {
				out <- GetItemsOutJob{
					Err: err,
					Id:  id,
				}

				continue
			}

			gzipEncodedData, err := ioutil.ReadAll(reader)
			if err != nil {
				out <- GetItemsOutJob{
					Err: err,
					Id:  id,
				}

				continue
			}

			out <- GetItemsOutJob{
				Err:             nil,
				Id:              id,
				GzipEncodedData: gzipEncodedData,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(16, worker, postWork)

	// enqueueing it up
	go func() {
		for _, id := range ids {
			in <- id
		}

		close(in)
	}()

	return out
}

func (b ItemsBase) WriteItem(obj *storage.ObjectHandle, item sotah.Item) error {
	jsonEncoded, err := json.Marshal(item)
	if err != nil {
		return err
	}

	gzipEncodedBody, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return err
	}

	wc := obj.NewWriter(b.client.Context)
	wc.ContentType = "application/json"
	wc.ContentEncoding = "gzip"
	return b.Write(wc, gzipEncodedBody)
}
