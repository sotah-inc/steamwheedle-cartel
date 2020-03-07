package store

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"

	"cloud.google.com/go/storage"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/store/regions"
)

func NewAreaMapsBase(c Client, location regions.Region, gameVersion gameversions.GameVersion) AreaMapsBase {
	return AreaMapsBase{base{client: c, location: location}, gameVersion}
}

type AreaMapsBase struct {
	base
	GameVersion gameversions.GameVersion
}

func (b AreaMapsBase) getBucketName() string {
	return "sotah-areamaps"
}

func (b AreaMapsBase) GetBucket() *storage.BucketHandle {
	return b.base.getBucket(b.getBucketName())
}

func (b AreaMapsBase) GetFirmBucket() (*storage.BucketHandle, error) {
	return b.base.getFirmBucket(b.getBucketName())
}

func (b AreaMapsBase) getObjectName(areaId int) string {
	return fmt.Sprintf("%s/%d.jpg", b.GameVersion, areaId)
}

func (b AreaMapsBase) GetObject(areaId int, bkt *storage.BucketHandle) *storage.ObjectHandle {
	return b.base.getObject(b.getObjectName(areaId), bkt)
}

func (b AreaMapsBase) GetFirmObject(areaId int, bkt *storage.BucketHandle) (*storage.ObjectHandle, error) {
	return b.base.getFirmObject(b.getObjectName(areaId), bkt)
}

func (b AreaMapsBase) WriteObject(areaId int, data []byte, bkt *storage.BucketHandle) error {
	wc := b.GetObject(areaId, bkt).NewWriter(b.client.Context)
	wc.ContentType = "image/jpeg"
	if _, err := wc.Write(data); err != nil {
		return err
	}

	return wc.Close()
}
