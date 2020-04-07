package diskstore

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (ds DiskStore) resolveItemFilepath(id blizzardv2.ItemId) string {
	return fmt.Sprintf("%s/items/%d.json", ds.CacheDir, id)
}
