package base

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func RetentionLimit() time.Time {
	return time.Now().Add(-1 * time.Hour * 24 * 1)
}

type databasePathPair struct {
	FullPath  string
	Timestamp sotah.UnixTimestamp
}

func Paths(databaseDir string) ([]databasePathPair, error) {
	databaseFilepaths, err := ioutil.ReadDir(databaseDir)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"dir":   databaseDir,
		}).Error("failed to read database dir")

		return []databasePathPair{}, err
	}

	out := make([]databasePathPair, len(databaseFilepaths))
	for i, fPath := range databaseFilepaths {
		targetTimestamp, err := strconv.Atoi(fPath.Name()[0 : len(fPath.Name())-len(".db")])
		if err != nil {
			logging.WithFields(logrus.Fields{
				"error":    err.Error(),
				"dir":      databaseDir,
				"pathname": fPath.Name(),
			}).Error("failed to parse database filepath")

			return []databasePathPair{}, err
		}

		fullPath, err := filepath.Abs(fmt.Sprintf("%s/%s", databaseDir, fPath.Name()))
		if err != nil {
			logging.WithFields(logrus.Fields{
				"error":    err.Error(),
				"dir":      databaseDir,
				"pathname": fPath.Name(),
			}).Error("failed to resolve full path of database file")

			return []databasePathPair{}, err
		}

		out[i] = databasePathPair{
			FullPath:  fullPath,
			Timestamp: sotah.UnixTimestamp(targetTimestamp),
		}
	}

	return out, nil
}
