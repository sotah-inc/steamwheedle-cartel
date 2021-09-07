package stats

import (
	"github.com/boltdb/bolt"
)

func newDatabase(dbFilepath string) (Database, error) {
	db, err := bolt.Open(dbFilepath, 0600, nil)
	if err != nil {
		return Database{}, err
	}

	return Database{db}, nil
}

type Database struct {
	db *bolt.DB
}
