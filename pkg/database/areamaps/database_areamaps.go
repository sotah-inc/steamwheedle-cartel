package areamaps

import (
	"fmt"
	"strconv"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

// bucketing
func baseBucketName() []byte {
	return []byte("area-maps")
}

func namesBucketName() []byte {
	return []byte("area-map-names")
}

// keying
func baseKeyName(id sotah.AreaMapId) []byte {
	return []byte(fmt.Sprintf("area-map-%d", id))
}

func nameKeyName(id sotah.AreaMapId) []byte {
	return []byte(fmt.Sprintf("area-map-name-%d", id))
}

func idFromNameKeyName(key []byte) (sotah.AreaMapId, error) {
	unparsedAreaMapId, err := strconv.Atoi(string(key)[len("area-map-name-"):])
	if err != nil {
		return 0, err
	}

	return sotah.AreaMapId(unparsedAreaMapId), nil
}

// db
func databasePath(dbDir string) (string, error) {
	return fmt.Sprintf("%s/areamaps.db", dbDir), nil
}
