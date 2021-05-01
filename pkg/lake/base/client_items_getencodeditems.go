package base

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemclass"
)

type GetEncodedItemJob interface {
	Err() error
	Id() blizzardv2.ItemId
	ItemClass() itemclass.Id
	EncodedItem() []byte
	EncodedNormalizedName() []byte
	IsVendorItem() bool
	VendorPrice() blizzardv2.PriceValue
	ToLogrusFields() logrus.Fields
}
