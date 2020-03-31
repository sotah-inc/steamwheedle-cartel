package diskstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func (ds DiskStore) resolveAuctionsFilepath(
	regionName blizzardv2.RegionName,
	realmSlug blizzardv2.RealmSlug,
) (string, error) {
	if len(ds.CacheDir) == 0 {
		return "", errors.New("cache dir cannot be blank")
	}

	return fmt.Sprintf("%s/auctions/%s/%s.json.gz", ds.CacheDir, regionName, realmSlug), nil
}

type WriteAuctionsOptions struct {
	RegionName blizzardv2.RegionName
	RealmSlug  blizzardv2.RealmSlug
	Auctions   blizzardv2.Auctions
}

func (ds DiskStore) WriteAuctions(opts WriteAuctionsOptions) error {
	dest, err := ds.resolveAuctionsFilepath(opts.RegionName, opts.RealmSlug)
	if err != nil {
		return err
	}

	jsonEncoded, err := json.Marshal(opts.Auctions)
	if err != nil {
		return err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return err
	}

	return util.WriteFile(dest, gzipEncoded)
}

func (ds DiskStore) GetAuctions(
	regionName blizzardv2.RegionName,
	realmSlug blizzardv2.RealmSlug,
) (blizzardv2.Auctions, time.Time, error) {
	// resolving the cached auctions filepath
	cachedAuctionsFilepath, err := ds.resolveAuctionsFilepath(regionName, realmSlug)
	if err != nil {
		return blizzardv2.Auctions{}, time.Time{}, err
	}

	// optionally skipping non-exist auctions file
	cachedAuctionsStat, err := os.Stat(cachedAuctionsFilepath)
	if err != nil {
		if !os.IsNotExist(err) {
			return blizzardv2.Auctions{}, time.Time{}, err
		}

		return blizzardv2.Auctions{}, time.Time{}, nil
	}

	gzipEncoded, err := util.ReadFile(cachedAuctionsFilepath)
	if err != nil {
		return blizzardv2.Auctions{}, time.Time{}, err
	}

	jsonEncoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return blizzardv2.Auctions{}, time.Time{}, err
	}

	var auctions blizzardv2.Auctions
	if err := json.Unmarshal(jsonEncoded, &auctions); err != nil {
		return blizzardv2.Auctions{}, time.Time{}, err
	}

	return auctions, cachedAuctionsStat.ModTime(), nil
}

type GetAuctionsByRealmsJob struct {
	Err          error
	Realm        sotah.Realm
	Auctions     blizzardv2.Auctions
	LastModified time.Time
}

func (job GetAuctionsByRealmsJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{}
}

func (ds DiskStore) GetAuctionsByRealms(reas sotah.Realms) chan GetAuctionsByRealmsJob {
	// establishing channels
	out := make(chan GetAuctionsByRealmsJob)
	in := make(chan sotah.Realm)

	// spinning up the workers for fetching auctions
	worker := func() {
		for rea := range in {
			aucs, lastModified, err := ds.GetAuctionsByRealm(rea)
			if lastModified.IsZero() {
				logging.WithFields(logrus.Fields{
					"region": rea.Region.Name,
					"realm":  rea.Slug,
				}).Error("Last-modified was blank when loading auctions from filecache")

				continue
			}

			out <- GetAuctionsByRealmsJob{err, rea, aucs, lastModified}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	// queueing up the realms
	go func() {
		for _, rea := range reas {
			logging.WithFields(logrus.Fields{
				"region": rea.Region.Name,
				"realm":  rea.Slug,
			}).Debug("Queueing up auction for loading")
			in <- rea
		}

		close(in)
	}()

	return out
}
