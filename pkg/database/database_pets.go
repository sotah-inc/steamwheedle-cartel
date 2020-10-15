package database

import (
	"fmt"
	"strconv"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

// bucketing
func databasePetsBucketName() []byte {
	return []byte("pets")
}

func databasePetNamesBucketName() []byte {
	return []byte("pet-names")
}

// keying
func petsKeyName(id blizzardv2.PetId) []byte {
	return []byte(fmt.Sprintf("pet-%d", id))
}

func petNameKeyName(id blizzardv2.PetId) []byte {
	return []byte(fmt.Sprintf("pet-name-%d", id))
}

func petIdFromPetKeyName(key []byte) (blizzardv2.PetId, error) {
	unparsedId, err := strconv.Atoi(string(key)[len("pet-"):])
	if err != nil {
		return blizzardv2.PetId(0), err
	}

	return blizzardv2.PetId(unparsedId), nil
}

func petIdFromPetNameKeyName(key []byte) (blizzardv2.PetId, error) {
	unparsedId, err := strconv.Atoi(string(key)[len("pet-name-"):])
	if err != nil {
		return blizzardv2.PetId(0), err
	}

	return blizzardv2.PetId(unparsedId), nil
}

// db
func PetsDatabasePath(dbDir string) (string, error) {
	return fmt.Sprintf("%s/pets.db", dbDir), nil
}
