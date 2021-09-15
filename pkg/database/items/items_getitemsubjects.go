package items

import (
	"errors"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

func (idBase Database) GetItemSubjects(
	version gameversion.GameVersion,
	ids blizzardv2.ItemIds,
) (blizzardv2.ItemSubjectsMap, error) {
	out := blizzardv2.ItemSubjectsMap{}
	for job := range idBase.FindItems(version, ids) {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("failed to find item")

			return blizzardv2.ItemSubjectsMap{}, job.Err
		}

		if !job.Exists {
			logging.WithField("id", job.Id).Error("attempted to fetch item not found")

			return blizzardv2.ItemSubjectsMap{}, errors.New("item not found")
		}

		foundName := job.Item.BlizzardMeta.Name.ResolveDefaultName()
		if foundName == "" {
			continue
		}

		out[job.Id] = foundName
	}

	return out, nil
}
