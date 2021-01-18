package store

import (
	"fmt"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/store/regions"
)

func NewPricelistHistoriesBaseV2(
	c Client,
	location regions.Region,
	version gameversions.GameVersion,
) PricelistHistoriesBaseV2 {
	return PricelistHistoriesBaseV2{
		base{client: c, location: location},
		version,
	}
}

type PricelistHistoriesBaseV2 struct {
	base
	GameVersion gameversions.GameVersion
}

func (b PricelistHistoriesBaseV2) getBucketName() string {
	return "sotah-pricelist-histories"
}

func (b PricelistHistoriesBaseV2) GetBucket() *storage.BucketHandle {
	return b.base.getBucket(b.getBucketName())
}

func (b PricelistHistoriesBaseV2) GetFirmBucket() (*storage.BucketHandle, error) {
	return b.base.getFirmBucket(b.getBucketName())
}

func (b PricelistHistoriesBaseV2) GetObjectPrefix(tuple blizzardv2.RegionConnectedRealmTuple) string {
	return fmt.Sprintf("%s/%s/%d", b.GameVersion, tuple.RegionName, tuple.ConnectedRealmId)
}

func (b PricelistHistoriesBaseV2) getObjectName(
	timestamp sotah.UnixTimestamp,
	tuple blizzardv2.RegionConnectedRealmTuple,
) string {
	return fmt.Sprintf("%s/%d.txt.gz", b.GetObjectPrefix(tuple), timestamp)
}

func (b PricelistHistoriesBaseV2) GetObject(
	timestamp sotah.UnixTimestamp,
	tuple blizzardv2.RegionConnectedRealmTuple,
	bkt *storage.BucketHandle,
) *storage.ObjectHandle {
	return b.base.getObject(b.getObjectName(timestamp, tuple), bkt)
}

func (b PricelistHistoriesBaseV2) GetFirmObject(
	timestamp sotah.UnixTimestamp,
	tuple blizzardv2.RegionConnectedRealmTuple,
	bkt *storage.BucketHandle,
) (*storage.ObjectHandle, error) {
	return b.base.getFirmObject(b.getObjectName(timestamp, tuple), bkt)
}

func (b PricelistHistoriesBaseV2) Handle(
	aucs blizzardv2.Auctions,
	targetTimestamp sotah.UnixTimestamp,
	tuple blizzardv2.RegionConnectedRealmTuple,
	bkt *storage.BucketHandle,
) (sotah.UnixTimestamp, error) {
	normalizedTargetDate := sotah.NormalizeToWeek(time.Unix(int64(targetTimestamp), 0))

	entry := logging.WithFields(logrus.Fields{
		"target-timestamp":       targetTimestamp,
		"normalized-target-date": normalizedTargetDate.Unix(),
	})

	// gathering an object
	entry.Info("gathering object")
	obj := b.GetObject(sotah.UnixTimestamp(normalizedTargetDate.Unix()), tuple, bkt)

	// resolving item-price-histories
	ipHistories, err := func() (sotah.ItemPriceHistories, error) {
		entry.Info("checking that object exists")
		exists, err := b.ObjectExists(obj)
		if err != nil {
			return sotah.ItemPriceHistories{}, err
		}

		if !exists {
			return sotah.ItemPriceHistories{}, nil
		}

		entry.Info("reading object")
		reader, err := obj.NewReader(b.client.Context)
		if err != nil {
			return sotah.ItemPriceHistories{}, err
		}
		defer func() {
			if err := reader.Close(); err != nil {
				entry.WithField("error", err.Error()).Error("failed to close reader")
			}
		}()

		return sotah.NewItemPriceHistoriesFromMinimized(reader)
	}()
	if err != nil {
		return 0, err
	}

	// gathering new item-prices from the input
	iPrices := sotah.NewItemPricesFromMiniAuctionList(sotah.NewMiniAuctionList(aucs))

	// merging item-prices into the item-price-histories
	for itemId, prices := range iPrices {
		pHistory := func() sotah.PriceHistory {
			result, ok := ipHistories[itemId]
			if !ok {
				return sotah.PriceHistory{}
			}

			return result
		}()
		pHistory[targetTimestamp] = prices

		ipHistories[itemId] = pHistory
	}

	// encoding the item-price-histories for persistence
	gzipEncodedBody, err := ipHistories.EncodeForPersistence()
	if err != nil {
		return 0, err
	}

	// writing it out to the gcloud object
	entry.Info("writing object")
	wc := obj.NewWriter(b.client.Context)
	wc.ContentType = "text/plain"
	wc.ContentEncoding = "gzip"
	if err := b.Write(wc, gzipEncodedBody); err != nil {
		entry.WithField("error", err.Error()).Error("failed to write object")

		return 0, err
	}

	return sotah.UnixTimestamp(normalizedTargetDate.Unix()), nil
}
