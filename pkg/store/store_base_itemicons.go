package store

import (
	"fmt"

	"cloud.google.com/go/storage"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah/gameversions"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store/regions"
)

const ItemIconURLFormat = "https://storage.googleapis.com/%s/%s"

func NewItemIconsBase(c Client, location regions.Region, version gameversions.GameVersion) ItemIconsBase {
	return ItemIconsBase{
		base{client: c, location: location},
		version,
	}
}

type ItemIconsBase struct {
	base
	GameVersion gameversions.GameVersion
}

func (b ItemIconsBase) GetBucketName() string {
	return "sotah-item-icons"
}

func (b ItemIconsBase) GetBucket() *storage.BucketHandle {
	return b.base.getBucket(b.GetBucketName())
}

func (b ItemIconsBase) GetFirmBucket() (*storage.BucketHandle, error) {
	return b.base.getFirmBucket(b.GetBucketName())
}

func (b ItemIconsBase) GetObjectName(name string) string {
	return fmt.Sprintf("%s/%s.jpg", b.GameVersion, name)
}

func (b ItemIconsBase) GetObject(name string, bkt *storage.BucketHandle) *storage.ObjectHandle {
	return b.base.getObject(b.GetObjectName(name), bkt)
}

func (b ItemIconsBase) GetFirmObject(name string, bkt *storage.BucketHandle) (*storage.ObjectHandle, error) {
	return b.base.getFirmObject(b.GetObjectName(name), bkt)
}
