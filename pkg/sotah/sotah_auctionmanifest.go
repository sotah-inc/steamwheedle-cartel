package sotah

import (
	"encoding/json"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
)

func NewAuctionManifestFromMap(am map[UnixTimestamp]interface{}) AuctionManifest {
	out := AuctionManifest{}
	for v := range am {
		out = append(out, v)
	}

	return out
}

type AuctionManifest []UnixTimestamp

func (am AuctionManifest) ToMap() map[UnixTimestamp]interface{} {
	out := map[UnixTimestamp]interface{}{}
	for _, v := range am {
		out[v] = struct{}{}
	}

	return out
}

func (am AuctionManifest) EncodeForPersistence() ([]byte, error) {
	jsonEncoded, err := json.Marshal(am)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncoded)
}

func (am AuctionManifest) Includes(subset AuctionManifest) bool {
	amMap := am.ToMap()
	subsetMap := subset.ToMap()
	for subsetTimestamp := range subsetMap {
		if _, ok := amMap[subsetTimestamp]; !ok {
			return false
		}
	}

	return true
}

func (am AuctionManifest) Merge(subset AuctionManifest) AuctionManifest {
	out := am.ToMap()
	for _, subsetTimestamp := range subset {
		out[subsetTimestamp] = struct{}{}
	}

	return NewAuctionManifestFromMap(out)
}

func NewCleanupManifestPayloadResponse(data string) (CleanupManifestPayloadResponse, error) {
	var out CleanupManifestPayloadResponse
	if err := json.Unmarshal([]byte(data), &out); err != nil {
		return CleanupManifestPayloadResponse{}, err
	}

	return out, nil
}

type CleanupManifestPayloadResponse struct {
	RegionRealmTuple
	TotalDeleted int `json:"total_removed"`
}

func (p CleanupManifestPayloadResponse) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	return string(jsonEncoded), nil
}
