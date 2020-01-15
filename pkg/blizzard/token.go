package blizzard

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const tokenInfoURLFormat = "https://%s/data/wow/token/index?namespace=dynamic-us"

func DefaultGetTokenInfoURL(regionHostname string) string {
	return fmt.Sprintf(tokenInfoURLFormat, regionHostname)
}

type GetTokenInfoURLFunc func(string) string

func NewTokenInfoFromHTTP(uri string) (TokenInfo, ResponseMeta, error) {
	resp, err := Download(uri)
	if err != nil {
		return TokenInfo{}, resp, err
	}

	if resp.Status != http.StatusOK {
		return TokenInfo{}, resp, errors.New("status was not 200")
	}

	tInfo, err := NewTokenInfo(resp.Body)
	if err != nil {
		return TokenInfo{}, resp, err
	}

	return tInfo, resp, nil
}

func NewTokenInfo(body []byte) (TokenInfo, error) {
	t := &TokenInfo{}
	if err := json.Unmarshal(body, t); err != nil {
		return TokenInfo{}, err
	}

	return *t, nil
}

type TokenInfo struct {
	LastUpdatedTimestamp int64 `json:"last_updated_timestamp"`
	Price                int64 `json:"price"`
}
