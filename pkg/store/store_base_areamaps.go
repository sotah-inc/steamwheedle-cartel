package store

import (
	"fmt"

	"cloud.google.com/go/storage"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/store/regions"
)

func NewAreaMapsBase(
	c Client,
	location regions.Region,
	version gameversions.GameVersion,
) AreaMapsBase {
	return AreaMapsBase{base{client: c, location: location}, version}
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

func (b AreaMapsBase) getObjectName(areaId sotah.AreaMapId) string {
	return fmt.Sprintf("%s/%d.jpg", b.GameVersion, areaId)
}

func (b AreaMapsBase) GetObject(
	areaId sotah.AreaMapId,
	bkt *storage.BucketHandle,
) *storage.ObjectHandle {
	return b.base.getObject(b.getObjectName(areaId), bkt)
}

func (b AreaMapsBase) GetFirmObject(
	areaId sotah.AreaMapId,
	bkt *storage.BucketHandle,
) (*storage.ObjectHandle, error) {
	return b.base.getFirmObject(b.getObjectName(areaId), bkt)
}

func (b AreaMapsBase) WriteObject(
	areaId sotah.AreaMapId,
	data []byte,
	bkt *storage.BucketHandle,
) error {
	wc := b.GetObject(areaId, bkt).NewWriter(b.client.Context)
	wc.ContentType = "image/jpeg"
	if _, err := wc.Write(data); err != nil {
		return err
	}

	return wc.Close()
}
