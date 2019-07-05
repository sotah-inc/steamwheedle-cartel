package database

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/database/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
)

func NewPricelistHistoryDatabases(dirPath string, statuses sotah.Statuses) (PricelistHistoryDatabases, error) {
	if len(dirPath) == 0 {
		return PricelistHistoryDatabases{}, errors.New("dir-path cannot be blank")
	}

	phdBases := PricelistHistoryDatabases{
		databaseDir: dirPath,
		Databases:   regionRealmDatabaseShards{},
	}

	for regionName, regionStatuses := range statuses {
		phdBases.Databases[regionName] = realmDatabaseShards{}

		for _, rea := range regionStatuses.Realms {
			phdBases.Databases[regionName][rea.Slug] = PricelistHistoryDatabaseShards{}

			dbPathPairs, err := Paths(fmt.Sprintf("%s/pricelist-histories/%s/%s", dirPath, regionName, rea.Slug))
			if err != nil {
				return PricelistHistoryDatabases{}, err
			}

			for _, dbPathPair := range dbPathPairs {
				phdBase, err := newPricelistHistoryDatabase(dbPathPair.FullPath, dbPathPair.TargetTime)
				if err != nil {
					return PricelistHistoryDatabases{}, err
				}

				phdBases.Databases[regionName][rea.Slug][sotah.UnixTimestamp(dbPathPair.TargetTime.Unix())] = phdBase
			}
		}
	}

	return phdBases, nil
}

type PricelistHistoryDatabases struct {
	databaseDir string
	Databases   regionRealmDatabaseShards
}

func (phdBases PricelistHistoryDatabases) resolveDatabaseFromLoadInJob(
	job LoadInJob,
) (PricelistHistoryDatabase, error) {
	normalizedTargetDate := sotah.NormalizeTargetDate(job.TargetTime)
	normalizedTargetTimestamp := sotah.UnixTimestamp(normalizedTargetDate.Unix())

	phdBase, ok := phdBases.Databases[job.Realm.Region.Name][job.Realm.Slug][normalizedTargetTimestamp]
	if ok {
		return phdBase, nil
	}

	dbPath := pricelistHistoryDatabaseFilePath(
		phdBases.databaseDir,
		job.Realm.Region.Name,
		job.Realm.Slug,
		normalizedTargetTimestamp,
	)
	phdBase, err := newPricelistHistoryDatabase(dbPath, normalizedTargetDate)
	if err != nil {
		return PricelistHistoryDatabase{}, err
	}
	phdBases.Databases[job.Realm.Region.Name][job.Realm.Slug][normalizedTargetTimestamp] = phdBase

	return phdBase, nil
}

type pricelistHistoriesLoadOutJob struct {
	Err          error
	Realm        sotah.Realm
	LastModified time.Time
}

func (job pricelistHistoriesLoadOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":         job.Err.Error(),
		"region":        job.Realm.Region.Name,
		"realm":         job.Realm.Slug,
		"last-modified": job.LastModified.Unix(),
	}
}

func (phdBases PricelistHistoryDatabases) Load(in chan LoadInJob) chan pricelistHistoriesLoadOutJob {
	// establishing channels
	out := make(chan pricelistHistoriesLoadOutJob)

	// spinning up workers for receiving auctions and persisting them
	worker := func() {
		for job := range in {
			phdBase, err := phdBases.resolveDatabaseFromLoadInJob(job)
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":  err.Error(),
					"region": job.Realm.Region.Name,
					"realm":  job.Realm.Slug,
				}).Error("Could not resolve database from load job")

				out <- pricelistHistoriesLoadOutJob{
					Err:          err,
					Realm:        job.Realm,
					LastModified: job.TargetTime,
				}

				continue
			}

			iPrices := sotah.NewItemPrices(sotah.NewMiniAuctionListFromMiniAuctions(sotah.NewMiniAuctions(job.Auctions)))
			if err := phdBase.persistItemPrices(job.TargetTime, iPrices); err != nil {
				logging.WithFields(logrus.Fields{
					"error":  err.Error(),
					"region": job.Realm.Region.Name,
					"realm":  job.Realm.Slug,
				}).Error("Failed to persist pricelists")

				out <- pricelistHistoriesLoadOutJob{
					Err:          err,
					Realm:        job.Realm,
					LastModified: job.TargetTime,
				}

				continue
			}

			out <- pricelistHistoriesLoadOutJob{
				Err:          nil,
				Realm:        job.Realm,
				LastModified: job.TargetTime,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(2, worker, postWork)

	return out
}

func (phdBases PricelistHistoryDatabases) pruneDatabases() error {
	earliestUnixTimestamp := RetentionLimit().Unix()
	logging.WithField("limit", earliestUnixTimestamp).Info("Checking for databases to prune")
	for rName, realmDatabases := range phdBases.Databases {
		for rSlug, databaseShards := range realmDatabases {
			for unixTimestamp, phdBase := range databaseShards {
				if int64(unixTimestamp) > earliestUnixTimestamp {
					continue
				}

				logging.WithFields(logrus.Fields{
					"region":             rName,
					"realm":              rSlug,
					"database-timestamp": unixTimestamp,
				}).Debug("Removing database from shard map")
				delete(phdBases.Databases[rName][rSlug], unixTimestamp)

				dbPath := phdBase.db.Path()

				logging.WithFields(logrus.Fields{
					"region":             rName,
					"realm":              rSlug,
					"database-timestamp": unixTimestamp,
				}).Debug("Closing database")
				if err := phdBase.db.Close(); err != nil {
					logging.WithFields(logrus.Fields{
						"region":   rName,
						"realm":    rSlug,
						"database": dbPath,
					}).Error("Failed to close database")

					return err
				}

				logging.WithFields(logrus.Fields{
					"region":   rName,
					"realm":    rSlug,
					"filepath": dbPath,
				}).Debug("Deleting database file")
				if err := os.Remove(dbPath); err != nil {
					logging.WithFields(logrus.Fields{
						"region":   rName,
						"realm":    rSlug,
						"database": dbPath,
					}).Error("Failed to remove database file")

					return err
				}
			}
		}
	}

	return nil
}

func (phdBases PricelistHistoryDatabases) StartPruner(stopChan sotah.WorkerStopChan) sotah.WorkerStopChan {
	onStop := make(sotah.WorkerStopChan)
	go func() {
		ticker := time.NewTicker(20 * time.Minute)

		logging.Info("Starting pruner")
	outer:
		for {
			select {
			case <-ticker.C:
				if err := phdBases.pruneDatabases(); err != nil {
					logging.WithField("error", err.Error()).Error("Failed to prune databases")

					continue
				}
			case <-stopChan:
				ticker.Stop()

				break outer
			}
		}

		onStop <- struct{}{}
	}()

	return onStop
}

func (phdBases PricelistHistoryDatabases) resolveDatabaseFromLoadInEncodedJob(
	job PricelistHistoryDatabaseEncodedLoadInJob,
) (PricelistHistoryDatabase, error) {
	phdBase, ok := phdBases.Databases[job.RegionName][job.RealmSlug][job.NormalizedTargetTimestamp]
	if ok {
		return phdBase, nil
	}

	normalizedTargetDate := time.Unix(int64(job.NormalizedTargetTimestamp), 0)

	dbPath := pricelistHistoryDatabaseFilePath(
		phdBases.databaseDir,
		job.RegionName,
		job.RealmSlug,
		job.NormalizedTargetTimestamp,
	)
	phdBase, err := newPricelistHistoryDatabase(dbPath, normalizedTargetDate)
	if err != nil {
		return PricelistHistoryDatabase{}, err
	}
	phdBases.Databases[job.RegionName][job.RealmSlug][job.NormalizedTargetTimestamp] = phdBase

	return phdBase, nil
}

func NewPricelistHistoriesComputeIntakeRequests(data string) (PricelistHistoriesComputeIntakeRequests, error) {
	base64Decoded, err := base64.RawStdEncoding.DecodeString(data)
	if err != nil {
		return PricelistHistoriesComputeIntakeRequests{}, err
	}

	gzipDecoded, err := util.GzipDecode(base64Decoded)
	if err != nil {
		return PricelistHistoriesComputeIntakeRequests{}, err
	}

	var out PricelistHistoriesComputeIntakeRequests
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return PricelistHistoriesComputeIntakeRequests{}, err
	}

	return out, nil
}

type PricelistHistoriesComputeIntakeRequests []PricelistHistoriesComputeIntakeRequest

func (r PricelistHistoriesComputeIntakeRequests) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(gzipEncoded), nil
}

type PricelistHistoriesComputeIntakeRequest struct {
	RegionName                string `json:"region_name"`
	RealmSlug                 string `json:"realm_slug"`
	NormalizedTargetTimestamp int    `json:"normalized_target_timestamp"`
}

func (r PricelistHistoriesComputeIntakeRequest) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	return string(jsonEncoded), nil
}

type PricelistHistoryDatabaseEncodedLoadInJob struct {
	RegionName                blizzard.RegionName
	RealmSlug                 blizzard.RealmSlug
	NormalizedTargetTimestamp sotah.UnixTimestamp
	VersionId                 string
	Data                      map[blizzard.ItemID][]byte
}

type PricelistHistoryDatabaseEncodedLoadOutJob struct {
	Err                       error
	RegionName                blizzard.RegionName
	RealmSlug                 blizzard.RealmSlug
	NormalizedTargetTimestamp sotah.UnixTimestamp
	VersionId                 string
}

func (job PricelistHistoryDatabaseEncodedLoadOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":                       job.Err.Error(),
		"region":                      job.RegionName,
		"realm":                       job.RealmSlug,
		"normalized-target-timestamp": job.NormalizedTargetTimestamp,
	}
}

func (phdBases PricelistHistoryDatabases) LoadEncoded(
	in chan PricelistHistoryDatabaseEncodedLoadInJob,
) chan PricelistHistoryDatabaseEncodedLoadOutJob {
	// establishing channels
	out := make(chan PricelistHistoryDatabaseEncodedLoadOutJob)

	// spinning up workers for receiving pre-encoded auctions and persisting them
	worker := func() {
		for job := range in {
			phdBase, err := phdBases.resolveDatabaseFromLoadInEncodedJob(job)
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":  err.Error(),
					"region": job.RegionName,
					"realm":  job.RealmSlug,
				}).Error("Could not resolve database from load job")

				out <- PricelistHistoryDatabaseEncodedLoadOutJob{
					Err:                       err,
					RegionName:                job.RegionName,
					RealmSlug:                 job.RealmSlug,
					NormalizedTargetTimestamp: job.NormalizedTargetTimestamp,
				}

				continue
			}

			if err := phdBase.persistEncodedItemPrices(job.Data); err != nil {
				logging.WithFields(logrus.Fields{
					"error":  err.Error(),
					"region": job.RegionName,
					"realm":  job.RealmSlug,
				}).Error("Could not persist encoded item-prices from job")

				out <- PricelistHistoryDatabaseEncodedLoadOutJob{
					Err:                       err,
					RegionName:                job.RegionName,
					RealmSlug:                 job.RealmSlug,
					NormalizedTargetTimestamp: job.NormalizedTargetTimestamp,
				}

				continue
			}

			out <- PricelistHistoryDatabaseEncodedLoadOutJob{
				Err:                       nil,
				RegionName:                job.RegionName,
				RealmSlug:                 job.RealmSlug,
				NormalizedTargetTimestamp: job.NormalizedTargetTimestamp,
				VersionId:                 job.VersionId,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(2, worker, postWork)

	return out
}

func NewGetPricelistHistoryRequest(data []byte) (GetPricelistHistoryRequest, error) {
	req := &GetPricelistHistoryRequest{}
	err := json.Unmarshal(data, &req)
	if err != nil {
		return GetPricelistHistoryRequest{}, err
	}

	return *req, nil
}

type GetPricelistHistoryRequest struct {
	RegionName  blizzard.RegionName `json:"region_name"`
	RealmSlug   blizzard.RealmSlug  `json:"realm_slug"`
	ItemIds     []blizzard.ItemID   `json:"item_ids"`
	LowerBounds int64               `json:"lower_bounds"`
	UpperBounds int64               `json:"upper_bounds"`
}

type GetPricelistHistoryResponse struct {
	History sotah.ItemPriceHistories `json:"history"`
}

func (res GetPricelistHistoryResponse) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(res)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}

func (phdBases PricelistHistoryDatabases) GetPricelistHistory(
	req GetPricelistHistoryRequest,
) (GetPricelistHistoryResponse, codes.Code, error) {
	regionShards, ok := phdBases.Databases[req.RegionName]
	if !ok {
		return GetPricelistHistoryResponse{}, codes.UserError, errors.New("invalid region")
	}

	realmShards, ok := regionShards[req.RealmSlug]
	if !ok {
		return GetPricelistHistoryResponse{}, codes.UserError, errors.New("invalid realm")
	}

	logging.WithFields(logrus.Fields{
		"shards": len(realmShards),
		"req":    fmt.Sprintf("+%v", req),
	}).Info("Querying shards")

	realm := sotah.NewSkeletonRealm(req.RegionName, req.RealmSlug)

	res := GetPricelistHistoryResponse{History: sotah.ItemPriceHistories{}}
	for _, ID := range req.ItemIds {
		plHistory, err := realmShards.GetPriceHistory(
			realm,
			ID,
			time.Unix(req.LowerBounds, 0),
			time.Unix(req.UpperBounds, 0),
		)
		if err != nil {
			return GetPricelistHistoryResponse{}, codes.GenericError, err
		}

		res.History[ID] = plHistory
	}

	return res, codes.Ok, nil
}
