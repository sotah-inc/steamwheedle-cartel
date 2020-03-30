package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NormalizeName(in string) (string, error) {
	return NormalizeString(in)
}

// item
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

type Item struct {
	BlizzardMeta blizzardv2.ItemResponse `json:"blizzard_meta"`
	SotahMeta    struct {
		IconURL        string         `json:"icon_url"`
		IconObjectName string         `json:"icon_object_name"`
		LastModified   int            `json:"last_modified"`
		NormalizedName locale.Mapping `json:"normalized_name"`
	} `json:"sotah_meta"`
}

func (item Item) EncodeForStorage() ([]byte, error) {
	jsonEncoded, err := json.Marshal(item)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncoded)
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

func NewItemIdNameMap(data []byte) (ItemIdNameMap, error) {
	gzipDecoded, err := util.GzipDecode(data)
	if err != nil {
		return ItemIdNameMap{}, err
	}

	var out ItemIdNameMap
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return ItemIdNameMap{}, err
	}

	return out, nil
}

type ItemIdNameMap map[blizzardv2.ItemId]locale.Mapping

func (idNameMap ItemIdNameMap) EncodeForDelivery() ([]byte, error) {
	jsonEncodedData, err := json.Marshal(idNameMap)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncodedData)
}

func (idNameMap ItemIdNameMap) ItemIds() []blizzardv2.ItemId {
	out := make([]blizzardv2.ItemId, len(idNameMap))
	i := 0
	for id := range idNameMap {
		out[i] = id

		i++
	}

	return out
}

func NewItemIdsBatches(ids []blizzardv2.ItemId, batchSize int) ItemIdBatches {
	batches := ItemIdBatches{}
	for i, id := range ids {
		key := (i - (i % batchSize)) / batchSize
		batch := func() []blizzardv2.ItemId {
			out, ok := batches[key]
			if !ok {
				return []blizzardv2.ItemId{}
			}

			return out
		}()
		batch = append(batch, id)

		batches[key] = batch
	}

	return batches
}

type ItemIdBatches map[int][]blizzardv2.ItemId

func NewIconItemsPayloadsBatches(iconIdsMap map[string][]blizzardv2.ItemId, batchSize int) IconItemsPayloadsBatches {
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
	Ids  []blizzardv2.ItemId
}
