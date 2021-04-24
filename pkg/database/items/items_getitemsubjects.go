package items

import (
	"errors"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

func (idBase Database) GetItemSubjects(ids blizzardv2.ItemIds) (blizzardv2.ItemSubjectsMap, error) {
	out := blizzardv2.ItemSubjectsMap{}
	for job := range idBase.FindItems(ids) {
		if job.Err != nil {
			return blizzardv2.ItemSubjectsMap{}, job.Err
		}

		if !job.Exists {
			logging.WithFields(job.ToLogrusFields()).Error("attempted to fetch item not found")

			return blizzardv2.ItemSubjectsMap{}, errors.New("item not found")
		}

		out[job.Id] = job.Item.SotahMeta.NormalizedName.ResolveDefaultName()
	}

	return out, nil
}
