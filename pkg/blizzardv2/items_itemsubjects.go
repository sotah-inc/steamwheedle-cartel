package blizzardv2

import (
	"encoding/base64"
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewItemSubjectsMap(base64Encoded string) (ItemSubjectsMap, error) {
	gzipEncoded, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return ItemSubjectsMap{}, err
	}

	jsonEncoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return ItemSubjectsMap{}, err
	}

	out := ItemSubjectsMap{}
	if err := json.Unmarshal(jsonEncoded, &out); err != nil {
		return ItemSubjectsMap{}, err
	}

	return out, nil
}

type ItemSubjectsMap map[ItemId]string

func (isMap ItemSubjectsMap) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(isMap)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}
