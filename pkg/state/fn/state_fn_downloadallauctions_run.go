package fn

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/act"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	mCodes "github.com/sotah-inc/steamwheedle-cartel/pkg/messenger/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/metric"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah/gameversions"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
)

func (sta DownloadAllAuctionsState) PublishToReceiveRealms(tuples sotah.RegionRealmTimestampTuples) error {
	// gathering a whitelist of region-realm-slugs
	regionRealmSlugs := tuples.ToRegionRealmSlugs()

	// gathering hell-realms for syncing
	hellRegionRealms, err := sta.IO.HellClient.GetRegionRealms(regionRealmSlugs, gameversions.Retail)
	if err != nil {
		return err
	}

	// updating the list of realms' timestamps
	for _, tuple := range tuples {
		hellRealm := hellRegionRealms[blizzard.RegionName(tuple.RegionName)][blizzard.RealmSlug(tuple.RealmSlug)]
		hellRealm.Downloaded = tuple.TargetTimestamp
		hellRegionRealms[blizzard.RegionName(tuple.RegionName)][blizzard.RealmSlug(tuple.RealmSlug)] = hellRealm

		logrus.WithFields(logrus.Fields{
			"region":     blizzard.RegionName(tuple.RegionName),
			"realm":      blizzard.RealmSlug(tuple.RealmSlug),
			"downloaded": tuple.TargetTimestamp,
		}).Info("Setting downloaded value for hell realm")
	}
	if err := sta.IO.HellClient.WriteRegionRealms(hellRegionRealms, gameversions.Retail); err != nil {
		return err
	}

	jsonEncoded, err := json.Marshal(regionRealmSlugs)
	if err != nil {
		return err
	}

	req, err := sta.IO.Messenger.Request(string(subjects.ReceiveRealms), jsonEncoded)
	if err != nil {
		return err
	}

	if req.Code != mCodes.Ok {
		return errors.New(req.Err)
	}

	return nil
}

func (sta DownloadAllAuctionsState) Run() error {
	regions, err := sta.bootBase.GetRegions(sta.bootBucket)
	if err != nil {
		return err
	}

	regionRealms := sotah.RegionRealms{}
	for _, region := range regions {
		realms, err := sta.realmsBase.GetAllRealms(region.Name, sta.realmsBucket)
		if err != nil {
			return err
		}

		regionRealms[region.Name] = realms
	}

	logging.WithField("endpoint-url", sta.actEndpoints.DownloadAuctions).Info("Producing act client")
	actClient, err := act.NewClient(sta.actEndpoints.DownloadAuctions)
	if err != nil {
		return err
	}

	logging.Info("Calling download-auctions with act client")
	actStartTime := time.Now()
	tuples := make(sotah.RegionRealmTimestampTuples, regionRealms.TotalRealms())
	for outJob := range actClient.DownloadAuctions(regionRealms) {
		// validating that no error occurred during act service calls
		if outJob.Err != nil {
			logging.WithFields(outJob.ToLogrusFields()).Error("Failed to fetch auctions")

			continue
		}

		// handling the job
		switch outJob.Data.Code {
		case http.StatusCreated:
			// parsing the response body
			tuple, err := sotah.NewRegionRealmTimestampTuple(string(outJob.Data.Body))
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":  err.Error(),
					"region": outJob.RegionName,
					"realm":  outJob.RealmSlug,
				}).Error("Failed to decode region-realm-timestamp tuple from act response body")

				break
			}

			tuples = append(tuples, tuple)
		case http.StatusNotModified:
			logging.WithFields(logrus.Fields{
				"region": outJob.RegionName,
				"realm":  outJob.RealmSlug,
			}).Info("Region-realm tuple was processed but no new auctions were found")
		default:
			logging.WithFields(logrus.Fields{
				"region":      outJob.RegionName,
				"realm":       outJob.RealmSlug,
				"status-code": outJob.Data.Code,
			}).Error("Response code for act call was not OK")
		}
	}
	durationInUs := int(int64(time.Since(actStartTime)) / 1000 / 1000 / 1000)
	logging.WithField(
		"duration-in-ms",
		durationInUs*1000,
	).Info("Finished calling act download-auctions")

	// reporting metrics
	sta.IO.Reporter.Report(metric.Metrics{
		"download_all_auctions_duration": int(int64(time.Since(actStartTime)) / 1000 / 1000 / 1000),
		"included_realms":                len(tuples),
	})

	// publishing to receive-realms
	logging.Info("Publishing realms to receive-realms")
	if err := sta.PublishToReceiveRealms(tuples); err != nil {
		return err
	}

	logging.Info("Finished")

	return nil
}
