package pets

import (
	"fmt"
	"strconv"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

// bucketing
func baseBucketName() []byte {
	return []byte("pets")
}

func namesBucketName() []byte {
	return []byte("pet-names")
}

func flagsBucketName() []byte {
	return []byte("flags")
}

// keying
func baseKeyName(id blizzardv2.PetId) []byte {
	return []byte(fmt.Sprintf("pet-%d", id))
}

func nameKeyName(id blizzardv2.PetId) []byte {
	return []byte(fmt.Sprintf("pet-name-%d", id))
}

func petIdFromBaseKeyName(key []byte) (blizzardv2.PetId, error) {
	unparsedId, err := strconv.Atoi(string(key)[len("pet-"):])
	if err != nil {
		return blizzardv2.PetId(0), err
	}

	return blizzardv2.PetId(unparsedId), nil
}

func petIdFromNameKeyName(key []byte) (blizzardv2.PetId, error) {
	unparsedId, err := strconv.Atoi(string(key)[len("pet-name-"):])
	if err != nil {
		return blizzardv2.PetId(0), err
	}

	return blizzardv2.PetId(unparsedId), nil
}

func isCompleteKeyName() []byte {
	return []byte("is-complete")
}

// db
func DatabasePath(dbDir string) (string, error) {
	return fmt.Sprintf("%s/pets.db", dbDir), nil
}
