package bus

import (
	"encoding/json"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func NewCleanupAuctionManifestJobs(regionExpiredTimestamps sotah.RegionRealmTimestamps) CleanupAuctionManifestJobs {
	out := CleanupAuctionManifestJobs{}
	for regionName, realmExpiredTimestamps := range regionExpiredTimestamps {
		for realmSlug, expiredTimestamps := range realmExpiredTimestamps {
			for _, timestamp := range expiredTimestamps {
				out = append(out, CleanupAuctionManifestJob{
					RegionName:      string(regionName),
					RealmSlug:       string(realmSlug),
					TargetTimestamp: int(timestamp),
				})
			}
		}
	}

	return out
}

type CleanupAuctionManifestJobs []CleanupAuctionManifestJob

func NewCleanupAuctionManifestJob(data string) (CleanupAuctionManifestJob, error) {
	var out CleanupAuctionManifestJob
	if err := json.Unmarshal([]byte(data), &out); err != nil {
		return CleanupAuctionManifestJob{}, err
	}

	return out, nil
}

type CleanupAuctionManifestJob struct {
	RegionName      string `json:"region_name"`
	RealmSlug       string `json:"realm_slug"`
	TargetTimestamp int    `json:"target_timestamp"`
}

func (c CleanupAuctionManifestJob) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	return string(jsonEncoded), nil
}

func NewCleanupAuctionManifestJobResponse(data string) (CleanupAuctionManifestJobResponse, error) {
	var out CleanupAuctionManifestJobResponse
	if err := json.Unmarshal([]byte(data), &out); err != nil {
		return CleanupAuctionManifestJobResponse{}, err
	}

	return out, nil
}

type CleanupAuctionManifestJobResponse struct {
	TotalDeleted int `json:"total_deleted"`
}

func (c CleanupAuctionManifestJobResponse) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	return string(jsonEncoded), nil
}
