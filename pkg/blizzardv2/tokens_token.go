package blizzardv2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const tokenInfoURLFormat = "https://%s/data/wow/token/index?namespace=dynamic-%s"

func DefaultGetTokenURL(regionHostname string, regionName RegionName) string {
	return fmt.Sprintf(tokenInfoURLFormat, regionHostname, regionName)
}

type GetTokenURLFunc func(string, RegionName) string

func NewTokenFromHTTP(uri string) (TokenResponse, ResponseMeta, error) {
	resp, err := Download(DownloadOptions{Uri: uri})
	if err != nil {
		return TokenResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		return TokenResponse{}, resp, errors.New("status was not 200")
	}

	tInfo, err := NewToken(resp.Body)
	if err != nil {
		return TokenResponse{}, resp, err
	}

	return tInfo, resp, nil
}

func NewToken(body []byte) (TokenResponse, error) {
	t := &TokenResponse{}
	if err := json.Unmarshal(body, t); err != nil {
		return TokenResponse{}, err
	}

	return *t, nil
}

type TokenResponse struct {
	LastUpdatedTimestamp int64 `json:"last_updated_timestamp"`
	Price                int64 `json:"price"`
}
