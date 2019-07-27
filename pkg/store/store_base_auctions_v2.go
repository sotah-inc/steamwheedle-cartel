package store

import (
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah/gameversions"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store/regions"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
	"google.golang.org/api/iterator"
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

func (b AuctionsBaseV2) GetObjectPrefix(realm sotah.Realm) string {
	return fmt.Sprintf("%s/%s/%s", b.GameVersion, realm.Region.Name, realm.Slug)
}

func (b AuctionsBaseV2) getObjectName(realm sotah.Realm, lastModified time.Time) string {
	return fmt.Sprintf("%s/%d.json.gz", b.GetObjectPrefix(realm), lastModified.Unix())
}

func (b AuctionsBaseV2) GetObject(
	realm sotah.Realm,
	lastModified time.Time,
	bkt *storage.BucketHandle,
) *storage.ObjectHandle {
	return b.base.getObject(b.getObjectName(realm, lastModified), bkt)
}

func (b AuctionsBaseV2) GetFirmObject(
	realm sotah.Realm,
	lastModified time.Time,
	bkt *storage.BucketHandle,
) (*storage.ObjectHandle, error) {
	return b.base.getFirmObject(b.getObjectName(realm, lastModified), bkt)
}

func (b AuctionsBaseV2) Handle(
	jsonEncodedBody []byte,
	lastModified time.Time,
	realm sotah.Realm,
	bkt *storage.BucketHandle,
) error {
	gzipEncodedBody, err := util.GzipEncode(jsonEncodedBody)
	if err != nil {
		return err
	}

	// writing it out to the gcloud object
	wc := b.GetObject(realm, lastModified, bkt).NewWriter(b.client.Context)
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

type DeleteAuctionsJob struct {
	Err             error
	TargetTimestamp sotah.UnixTimestamp
}

func (b AuctionsBaseV2) DeleteAll(
	bkt *storage.BucketHandle,
	realm sotah.Realm,
	manifest sotah.AuctionManifest,
) chan DeleteAuctionsJob {
	// spinning up the workers
	in := make(chan sotah.UnixTimestamp)
	out := make(chan DeleteAuctionsJob)
	worker := func() {
		for targetTimestamp := range in {
			obj := bkt.Object(b.getObjectName(realm, time.Unix(int64(targetTimestamp), 0)))
			exists, err := b.ObjectExists(obj)
			if err != nil {
				out <- DeleteAuctionsJob{
					Err:             err,
					TargetTimestamp: targetTimestamp,
				}

				continue
			}
			if !exists {
				out <- DeleteAuctionsJob{
					Err:             nil,
					TargetTimestamp: targetTimestamp,
				}

				continue
			}

			if err := obj.Delete(b.client.Context); err != nil {
				out <- DeleteAuctionsJob{
					Err:             err,
					TargetTimestamp: targetTimestamp,
				}

				continue
			}

			out <- DeleteAuctionsJob{
				Err:             nil,
				TargetTimestamp: targetTimestamp,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(16, worker, postWork)

	// queueing it up
	go func() {
		for _, targetTimestamp := range manifest {
			in <- targetTimestamp
		}

		close(in)
	}()

	return out
}

func (b AuctionsBaseV2) GetExpiredTimestamps(
	realm sotah.Realm,
	bkt *storage.BucketHandle,
) ([]sotah.UnixTimestamp, error) {
	timestamps, err := b.GetTimestamps(realm, bkt)
	if err != nil {
		return []sotah.UnixTimestamp{}, err
	}

	limit := sotah.NormalizeTargetDate(time.Now()).AddDate(0, 0, -14)
	expiredTimestamps := []sotah.UnixTimestamp{}
	for _, timestamp := range timestamps {
		targetTime := time.Unix(int64(timestamp), 0)
		if targetTime.After(limit) {
			continue
		}

		expiredTimestamps = append(expiredTimestamps, timestamp)
	}

	return expiredTimestamps, nil
}

func (b AuctionsBaseV2) GetTimestamps(
	realm sotah.Realm,
	bkt *storage.BucketHandle,
) ([]sotah.UnixTimestamp, error) {
	prefix := fmt.Sprintf("%s/", b.GetObjectPrefix(realm))
	it := bkt.Objects(b.client.Context, &storage.Query{Prefix: prefix})
	out := []sotah.UnixTimestamp{}
	for {
		objAttrs, err := it.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}

			return []sotah.UnixTimestamp{}, err
		}

		targetTimestamp, err := strconv.Atoi(objAttrs.Name[len(prefix):(len(objAttrs.Name) - len(".json.gz"))])
		if err != nil {
			return []sotah.UnixTimestamp{}, err
		}

		out = append(out, sotah.UnixTimestamp(targetTimestamp))
	}

	return out, nil
}

func (b AuctionsBaseV2) DeleteAllFromTimestamps(
	timestamps []sotah.UnixTimestamp,
	realm sotah.Realm,
	bkt *storage.BucketHandle,
) (DeleteAllResults, error) {
	// spinning up the workers
	in := make(chan sotah.UnixTimestamp)
	out := make(chan DeleteAllFromTimestampsJob)
	worker := func() {
		for targetTimestamp := range in {
			entry := logging.WithFields(logrus.Fields{
				"region":           realm.Region.Name,
				"realm":            realm.Slug,
				"target-timestamp": targetTimestamp,
			})
			entry.Info("Handling target-timestamp")

			obj := bkt.Object(b.getObjectName(realm, time.Unix(int64(targetTimestamp), 0)))

			exists, err := b.ObjectExists(obj)
			if err != nil {
				entry.WithField("error", err.Error()).Error("Failed to check if obj exists")

				out <- DeleteAllFromTimestampsJob{
					Err: err,
					RegionRealmTimestampTuple: sotah.RegionRealmTimestampTuple{
						RegionRealmTuple: sotah.NewRegionRealmTupleFromRealm(realm),
						TargetTimestamp:  int(targetTimestamp),
					},
				}

				continue
			}
			if !exists {
				entry.Info("Obj does not exist")

				out <- DeleteAllFromTimestampsJob{
					Err: nil,
					RegionRealmTimestampTuple: sotah.RegionRealmTimestampTuple{
						RegionRealmTuple: sotah.NewRegionRealmTupleFromRealm(realm),
						TargetTimestamp:  int(targetTimestamp),
					},
				}

				continue
			}

			attrs, err := obj.Attrs(b.client.Context)
			if err != nil {
				entry.WithField("error", err.Error()).Error("Failed to get obj attrs")

				out <- DeleteAllFromTimestampsJob{
					Err: err,
					RegionRealmTimestampTuple: sotah.RegionRealmTimestampTuple{
						RegionRealmTuple: sotah.NewRegionRealmTupleFromRealm(realm),
						TargetTimestamp:  int(targetTimestamp),
					},
				}

				continue
			}

			if err := obj.Delete(b.client.Context); err != nil {
				entry.WithField("error", err.Error()).Error("Could not delete obj")

				out <- DeleteAllFromTimestampsJob{
					Err: err,
					RegionRealmTimestampTuple: sotah.RegionRealmTimestampTuple{
						RegionRealmTuple: sotah.NewRegionRealmTupleFromRealm(realm),
						TargetTimestamp:  int(targetTimestamp),
					},
				}

				continue
			}

			entry.Info("Obj deleted")

			out <- DeleteAllFromTimestampsJob{
				RegionRealmTimestampTuple: sotah.RegionRealmTimestampTuple{
					RegionRealmTuple: sotah.NewRegionRealmTupleFromRealm(realm),
					TargetTimestamp:  int(targetTimestamp),
				},
				Err:  nil,
				Size: attrs.Size,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(16, worker, postWork)

	// queueing it up
	go func() {
		for _, targetTimestamp := range timestamps {
			in <- targetTimestamp
		}

		close(in)
	}()

	// waiting for it to drain out
	results := DeleteAllResults{
		TotalCount: 0,
		TotalSize:  0,
	}
	for outJob := range out {
		if outJob.Err != nil {
			return DeleteAllResults{}, outJob.Err
		}

		results.TotalCount += 1
		results.TotalSize += outJob.Size
	}

	return results, nil
}
