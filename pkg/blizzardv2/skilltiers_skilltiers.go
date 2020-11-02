package blizzardv2

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetAllSkillTiersOptions struct {
	GetSkillTierURL func(SkillTierId) (string, error)

	SkillTierIdList []SkillTierId
	Blacklist       []SkillTierId
}

type GetAllSkillTiersJob struct {
	Err               error
	Id                SkillTierId
	SkillTierResponse SkillTierResponse
}

func (job GetAllSkillTiersJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"id":    job.Id,
	}
}

func GetAllSkillTiers(opts GetAllSkillTiersOptions) (chan GetAllSkillTiersJob, error) {
	// starting up workers for gathering individual skillTiers
	in := make(chan SkillTierId)
	out := make(chan GetAllSkillTiersJob)
	worker := func() {
		for id := range in {
			GetSkillTierUri, err := opts.GetSkillTierURL(id)
			if err != nil {
				out <- GetAllSkillTiersJob{
					Err:               err,
					Id:                id,
					SkillTierResponse: SkillTierResponse{},
				}

				continue
			}

			skillTier, _, err := NewSkillTierResponseFromHTTP(GetSkillTierUri)
			if err != nil {
				out <- GetAllSkillTiersJob{
					Err:               err,
					Id:                id,
					SkillTierResponse: SkillTierResponse{},
				}

				continue
			}

			out <- GetAllSkillTiersJob{
				Err:               nil,
				Id:                id,
				SkillTierResponse: skillTier,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	// converting blacklist to map for filtering
	blacklistMap := map[SkillTierId]struct{}{}
	for _, id := range opts.Blacklist {
		blacklistMap[id] = struct{}{}
	}

	// queueing it up
	go func() {
		total := 0
		for _, id := range opts.SkillTierIdList {
			if _, ok := blacklistMap[id]; ok {
				continue
			}

			in <- id

			total += 1
		}

		close(in)
	}()

	return out, nil
}
