package blizzardv2

import "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"

type ItemId int

type ItemQuality struct {
	Type string         `json:"type"`
	Name locale.Mapping `json:"name"`
}

type ItemMedia struct {
	Key HrefReference `json:"key"`
	Id  ItemId        `json:"id"`
}

type ItemInventoryType struct {
	Type string         `json:"type"`
	Name locale.Mapping `json:"name"`
}

type ItemSpell struct {
	Spell struct {
		Key  HrefReference  `json:"key"`
		Name locale.Mapping `json:"name"`
		Id   int            `json:"id"`
	} `json:"spell"`
	Description locale.Mapping `json:"description"`
}

type Item struct {
	LinksBase
	Id            ItemId            `json:"id"`
	Name          locale.Mapping    `json:"name"`
	Quality       ItemQuality       `json:"quality"`
	Level         int               `json:"level"`
	RequiredLevel int               `json:"required_level"`
	Media         ItemMedia         `json:"media"`
	ItemClass     ItemClass         `json:"item_class"`
	ItemSubClass  ItemSubClass      `json:"item_subclass"`
	InventoryType ItemInventoryType `json:"inventory_type"`
	PurchasePrice int64             `json:"purchase_price"`
	SellPrice     int64             `json:"sell_price"`
	MaxCount      int               `json:"max_count"`
	IsEquippable  bool              `json:"is_equippable"`
	IsStackable   bool              `json:"is_stackable"`
	PreviewItem   struct {
		Item struct {
			Key HrefReference `json:"key"`
			Id  ItemId        `json:"id"`
		} `json:"item"`
		Quality       ItemQuality       `json:"quality"`
		Name          locale.Mapping    `json:"name"`
		Media         ItemMedia         `json:"media"`
		ItemClass     ItemClass         `json:"item_class"`
		ItemSubClass  ItemSubClass      `json:"item_subclass"`
		InventoryType ItemInventoryType `json:"inventory_type"`
		Binding       struct {
			Type string         `json:"type"`
			Name locale.Mapping `json:"name"`
		} `json:"binding"`
		Spells    []ItemSpell `json:"spells"`
		SellPrice struct {
			Value  int64          `json:"value"`
			Header locale.Mapping `json:"header"`
			Gold   locale.Mapping `json:"gold"`
			Silver locale.Mapping `json:"silver"`
			Copper locale.Mapping `json:"copper"`
		} `json:"sell_price"`
		IsSubClassHidden bool `json:"is_subclass_hidden"`
		NameDescription  struct {
			DisplayString locale.Mapping `json:"display_string"`
			Color         struct {
				Red   int     `json:"r"`
				Green int     `json:"g"`
				Blue  int     `json:"b"`
				Alpha float32 `json:"a"`
			} `json:"color"`
		} `json:"name_description"`
	} `json:"preview_item"`
}
