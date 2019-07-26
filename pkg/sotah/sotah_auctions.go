package sotah

import "encoding/json"

func NewCleanupAuctionsPayloadResponse(data string) (CleanupAuctionsPayloadResponse, error) {
	var out CleanupAuctionsPayloadResponse
	if err := json.Unmarshal([]byte(data), &out); err != nil {
		return CleanupAuctionsPayloadResponse{}, err
	}

	return out, nil
}

type CleanupAuctionsPayloadResponse struct {
	RegionRealmTuple
	TotalDeleted int `json:"total_removed"`
}

func (p CleanupAuctionsPayloadResponse) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	return string(jsonEncoded), nil
}
