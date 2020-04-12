package store

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	"cloud.google.com/go/storage"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/store/regions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewAuctionsBaseV2(c Client, location regions.Region, version gameversions.GameVersion) AuctionsBaseV2 {
	return AuctionsBaseV2{
		base{client: c, location: location},
		version,
	}
}

type AuctionsBaseV2 struct {
	base
	GameVersion gameversions.GameVersion
}

func (b AuctionsBaseV2) getBucketName() string {
	return "sotah-raw-auctions"
}

func (b AuctionsBaseV2) GetBucket() *storage.BucketHandle {
	return b.base.getBucket(b.getBucketName())
}

func (b AuctionsBaseV2) GetFirmBucket() (*storage.BucketHandle, error) {
	return b.base.getFirmBucket(b.getBucketName())
}

func (b AuctionsBaseV2) ResolveBucket() (*storage.BucketHandle, error) {
	return b.base.resolveBucket(b.getBucketName())
}

func (b AuctionsBaseV2) GetObjectPrefix(tuple blizzardv2.RegionConnectedRealmTuple) string {
	return fmt.Sprintf("%s/%s/%d", b.GameVersion, tuple.RegionName, tuple.ConnectedRealmId)
}

func (b AuctionsBaseV2) getObjectName(
	tuple blizzardv2.RegionConnectedRealmTuple,
	timestamp sotah.UnixTimestamp,
) string {
	return fmt.Sprintf("%s/%d.json.gz", b.GetObjectPrefix(tuple), timestamp)
}

func (b AuctionsBaseV2) GetObject(
	tuple blizzardv2.RegionConnectedRealmTuple,
	timestamp sotah.UnixTimestamp,
	bkt *storage.BucketHandle,
) *storage.ObjectHandle {
	return b.base.getObject(b.getObjectName(tuple, timestamp), bkt)
}

func (b AuctionsBaseV2) GetFirmObject(
	tuple blizzardv2.RegionConnectedRealmTuple,
	timestamp sotah.UnixTimestamp,
	bkt *storage.BucketHandle,
) (*storage.ObjectHandle, error) {
	return b.base.getFirmObject(b.getObjectName(tuple, timestamp), bkt)
}

func (b AuctionsBaseV2) Handle(
	jsonEncodedBody []byte,
	timestamp sotah.UnixTimestamp,
	tuple blizzardv2.RegionConnectedRealmTuple,
	bkt *storage.BucketHandle,
) error {
	gzipEncodedBody, err := util.GzipEncode(jsonEncodedBody)
	if err != nil {
		return err
	}

	// writing it out to the gcloud object
	wc := b.GetObject(tuple, timestamp, bkt).NewWriter(b.client.Context)
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
