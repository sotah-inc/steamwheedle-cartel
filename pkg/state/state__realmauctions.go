package state

import (
	"encoding/json"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type StoreAuctionsInJob struct {
	Realm      sotah.Realm
	TargetTime time.Time
	Auctions   blizzard.Auctions
}

type StoreAuctionsOutJob struct {
	Err        error
	Realm      sotah.Realm
	TargetTime time.Time
	ItemIds    []blizzardv2.ItemId
}

func (job StoreAuctionsOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":       job.Err.Error(),
		"region":      job.Realm.Region.Name,
		"realm":       job.Realm.Slug,
		"target-time": job.TargetTime.Unix(),
	}
}

func (sta State) StoreAuctions(in chan StoreAuctionsInJob) chan StoreAuctionsOutJob {
	out := make(chan StoreAuctionsOutJob)

	// spinning up the workers for fetching Auctions
	worker := func() {
		for inJob := range in {
			jsonEncodedData, err := json.Marshal(inJob.Auctions)
			if err != nil {
				out <- StoreAuctionsOutJob{
					Err:        err,
					Realm:      inJob.Realm,
					TargetTime: inJob.TargetTime,
					ItemIds:    []blizzardv2.ItemId{},
				}

				continue
			}

			gzipEncodedData, err := util.GzipEncode(jsonEncodedData)
			if err != nil {
				out <- StoreAuctionsOutJob{
					Err:        err,
					Realm:      inJob.Realm,
					TargetTime: inJob.TargetTime,
					ItemIds:    []blizzardv2.ItemId{},
				}

				continue
			}

			if err := sta.IO.DiskStore.WriteAuctions(inJob.Realm, gzipEncodedData); err != nil {
				out <- StoreAuctionsOutJob{
					Err:        err,
					Realm:      inJob.Realm,
					TargetTime: inJob.TargetTime,
					ItemIds:    []blizzardv2.ItemId{},
				}

				continue
			}

			out <- StoreAuctionsOutJob{
				Err:        nil,
				Realm:      inJob.Realm,
				TargetTime: inJob.TargetTime,
				ItemIds:    inJob.Auctions.ItemIds(),
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	return out
}
