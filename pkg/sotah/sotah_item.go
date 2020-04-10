package sotah

import (
	"encoding/json"
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type IconName string

func NewItemObjectName(name IconName) string {
	return fmt.Sprintf(
		"%s/%s.jpg",
		gameversions.Retail,
		name,
	)
}

type ItemIconMeta struct {
	URL        string   `json:"icon_url"`
	ObjectName string   `json:"icon_object_name"`
	Icon       IconName `json:"icon"`
}

func (meta ItemIconMeta) IsZero() bool {
	return meta.URL == "" || meta.ObjectName == "" || meta.Icon == ""
}

func NewItemFromGzipped(gzipEncoded []byte) (Item, error) {
	gzipDecoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return Item{}, err
	}

	return NewItem(gzipDecoded)
}

func NewItem(body []byte) (Item, error) {
	i := &Item{}
	if err := json.Unmarshal(body, i); err != nil {
		return Item{}, err
	}

	return *i, nil
}

type ItemMeta struct {
	LastModified   UnixTime       `json:"last_modified"`
	NormalizedName locale.Mapping `json:"normalized_name"`
	ItemIconMeta   ItemIconMeta   `json:"item_icon_meta"`
}

type Item struct {
	BlizzardMeta blizzardv2.ItemResponse `json:"blizzard_meta"`
	SotahMeta    ItemMeta                `json:"sotah_meta"`
}

func (item Item) EncodeForStorage() ([]byte, error) {
	jsonEncoded, err := json.Marshal(item)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncoded)
}
