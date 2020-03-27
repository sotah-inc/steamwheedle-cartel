package sotah

import (
	"encoding/base64"
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NormalizeName(in string) (string, error) {
	return NormalizeString(in)
}

// item
func NewItem(body []byte) (Item, error) {
	i := &Item{}
	if err := json.Unmarshal(body, i); err != nil {
		return Item{}, err
	}

	return *i, nil
}

type Item struct {
	BlizzardMeta blizzardv2.ItemResponse `json:"blizzard_meta"`
	SotahMeta    struct {
		IconURL        string `json:"icon_url"`
		IconObjectName string `json:"icon_object_name"`
		LastModified   int    `json:"last_modified"`
	} `json:"sotah_meta"`
}

// item-icon-item-ids map
type ItemIconItemIdsMap map[string][]blizzardv2.ItemId

func (iconsMap ItemIconItemIdsMap) GetItemIcons() []string {
	iconNames := make([]string, len(iconsMap))
	i := 0
	for iconName := range iconsMap {
		iconNames[i] = iconName

		i++
	}

	return iconNames
}

// item-ids map
type ItemIdsMap map[blizzardv2.ItemId]struct{}

// items-map
type ItemsMap map[blizzardv2.ItemId]Item

func (iMap ItemsMap) GetItemIconsMap(excludeWithURL bool) ItemIconItemIdsMap {
	out := ItemIconItemIdsMap{}
	for itemId, iValue := range iMap {
		if excludeWithURL && iValue.SotahMeta.IconURL != "" {
			continue
		}

		if iValue.ItemResponse == "" {
			continue
		}

		if _, ok := out[iValue.Icon]; !ok {
			out[iValue.Icon] = []blizzard.ItemID{itemId}

			continue
		}

		out[iValue.Icon] = append(out[iValue.Icon], itemId)
	}

	return out
}

func (iMap ItemsMap) EncodeForDatabase() ([]byte, error) {
	jsonEncodedData, err := json.Marshal(iMap)
	if err != nil {
		return []byte{}, err
	}

	gzipEncodedData, err := util.GzipEncode(jsonEncodedData)
	if err != nil {
		return []byte{}, err
	}

	return gzipEncodedData, nil
}

func NewItemIdNameMap(data string) (ItemIdNameMap, error) {
	base64Decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return ItemIdNameMap{}, err
	}

	gzipDecoded, err := util.GzipDecode(base64Decoded)
	if err != nil {
		return ItemIdNameMap{}, err
	}

	var out ItemIdNameMap
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return ItemIdNameMap{}, err
	}

	return out, nil
}

type ItemIdNameMap map[blizzard.ItemID]string

func (idNameMap ItemIdNameMap) EncodeForDelivery() (string, error) {
	jsonEncodedData, err := json.Marshal(idNameMap)
	if err != nil {
		return "", err
	}

	gzipEncodedData, err := util.GzipEncode(jsonEncodedData)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncodedData), nil
}

func (idNameMap ItemIdNameMap) ItemIds() blizzard.ItemIds {
	out := make(blizzard.ItemIds, len(idNameMap))
	i := 0
	for id := range idNameMap {
		out[i] = id

		i++
	}

	return out
}

func NewItemIdsBatches(ids blizzard.ItemIds, batchSize int) ItemIdBatches {
	batches := ItemIdBatches{}
	for i, id := range ids {
		key := (i - (i % batchSize)) / batchSize
		batch := func() blizzard.ItemIds {
			out, ok := batches[key]
			if !ok {
				return blizzard.ItemIds{}
			}

			return out
		}()
		batch = append(batch, id)

		batches[key] = batch
	}

	return batches
}

type ItemIdBatches map[int]blizzard.ItemIds

func NewIconItemsPayloadsBatches(iconIdsMap map[string]blizzard.ItemIds, batchSize int) IconItemsPayloadsBatches {
	batches := IconItemsPayloadsBatches{}
	i := 0
	for iconName, itemIds := range iconIdsMap {
		key := (i - (i % batchSize)) / batchSize
		batch := func() IconItemsPayloads {
			out, ok := batches[key]
			if !ok {
				return IconItemsPayloads{}
			}

			return out
		}()
		batch = append(batch, IconItemsPayload{Name: iconName, Ids: itemIds})

		batches[key] = batch

		i += 1
	}

	return batches
}

type IconItemsPayloadsBatches map[int]IconItemsPayloads

func NewIconItemsPayloads(data string) (IconItemsPayloads, error) {
	var out IconItemsPayloads
	if err := json.Unmarshal([]byte(data), &out); err != nil {
		return IconItemsPayloads{}, err
	}

	return out, nil
}

type IconItemsPayloads []IconItemsPayload

func (d IconItemsPayloads) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(d)
	if err != nil {
		return "", err
	}

	return string(jsonEncoded), nil
}

type IconItemsPayload struct {
	Name string
	Ids  blizzard.ItemIds
}
