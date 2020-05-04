package base

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type Client interface {
	Collect() (blizzardv2.ItemIds, error)
}
