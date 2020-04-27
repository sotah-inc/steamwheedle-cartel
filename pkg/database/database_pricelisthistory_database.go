package database

import (
	"time"

	"github.com/boltdb/bolt"
)

type PricelistHistoryDatabase struct {
	db         *bolt.DB
	targetDate time.Time
}
