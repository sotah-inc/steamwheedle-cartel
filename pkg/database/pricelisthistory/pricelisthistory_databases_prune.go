package pricelisthistory

import (
	"os"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (phdBases *Databases) PruneDatabases(retentionLimitTimestamp sotah.UnixTimestamp) error {
	logging.WithFields(logrus.Fields{
		"limit": retentionLimitTimestamp,
		"total": phdBases.Total(),
	}).Info("checking for databases to prune")

	for rName, versionDatabases := range phdBases.Databases {
		for vName, realmDatabases := range versionDatabases {
			for rSlug, databaseShards := range realmDatabases {
				for unixTimestamp, phdBase := range databaseShards.Before(retentionLimitTimestamp, false) {
					logging.WithFields(logrus.Fields{
						"region":             rName,
						"game-version":       vName,
						"realm":              rSlug,
						"database-timestamp": unixTimestamp,
					}).Debug("removing database from shard map")
					delete(phdBases.Databases[rName][vName][rSlug], unixTimestamp)

					dbPath := phdBase.db.Path()

					logging.WithFields(logrus.Fields{
						"region":             rName,
						"realm":              rSlug,
						"database-timestamp": unixTimestamp,
					}).Debug("closing database")
					if err := phdBase.db.Close(); err != nil {
						logging.WithFields(logrus.Fields{
							"region":   rName,
							"realm":    rSlug,
							"database": dbPath,
						}).Error("failed to close database")

						return err
					}

					logging.WithFields(logrus.Fields{
						"region":   rName,
						"realm":    rSlug,
						"filepath": dbPath,
					}).Debug("deleting database file")
					if err := os.Remove(dbPath); err != nil {
						logging.WithFields(logrus.Fields{
							"region":   rName,
							"realm":    rSlug,
							"database": dbPath,
						}).Error("failed to remove database file")

						return err
					}
				}
			}
		}
	}

	logging.WithFields(logrus.Fields{
		"limit": retentionLimitTimestamp,
		"total": phdBases.Total(),
	}).Info("done checking for databases to prune")

	return nil
}
