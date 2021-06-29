package store

import (
	"fmt"

	"cloud.google.com/go/storage"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/store/regions"
)

func NewLiveAuctionsBase(
	c Client,
	location regions.Region,
	version gameversions.GameVersion,
) LiveAuctionsBase {
	return LiveAuctionsBase{
		base{client: c, location: location},
		version,
	}
}

type LiveAuctionsBase struct {
	base
	GameVersion gameversions.GameVersion
}

func (b LiveAuctionsBase) getBucketName() string {
	return "sotah-live-auctions"
}

func (b LiveAuctionsBase) GetFirmBucket() (*storage.BucketHandle, error) {
	return b.base.getFirmBucket(b.getBucketName())
}

func (b LiveAuctionsBase) GetBucket() *storage.BucketHandle {
	return b.base.getBucket(b.getBucketName())
}

func (b LiveAuctionsBase) getObjectName(tuple blizzardv2.RegionVersionConnectedRealmTuple) string {
	return fmt.Sprintf("%s/%s/%d.json.gz", b.GameVersion, tuple.RegionName, tuple.ConnectedRealmId)
}

func (b LiveAuctionsBase) GetObject(
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
	bkt *storage.BucketHandle,
) *storage.ObjectHandle {
	return b.base.getObject(b.getObjectName(tuple), bkt)
}

func (b LiveAuctionsBase) GetFirmObject(
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
	bkt *storage.BucketHandle,
) (*storage.ObjectHandle, error) {
	return b.base.getFirmObject(b.getObjectName(tuple), bkt)
}

func (b LiveAuctionsBase) Handle(
	aucs blizzardv2.Auctions,
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
	bkt *storage.BucketHandle,
) error {
	// encoding auctions in the appropriate format
	gzipEncodedBody, err := sotah.NewMiniAuctionList(aucs).EncodeForStorage()
	if err != nil {
		return err
	}

	// writing it out to the gcloud object
	wc := b.GetObject(tuple, bkt).NewWriter(b.client.Context)
	wc.ContentType = "application/json"
	wc.ContentEncoding = "gzip"
	if err := b.Write(wc, gzipEncodedBody); err != nil {
		return err
	}

	return nil
}
