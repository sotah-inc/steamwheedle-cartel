package base

import "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"

type Client interface {
	Collect(version gameversion.GameVersion) error
}
