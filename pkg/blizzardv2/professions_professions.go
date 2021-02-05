package blizzardv2

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetAllProfessionsOptions struct {
	GetProfessionIndexURL func() (string, error)
	GetProfessionURL      func(string) (string, error)

	Blacklist []ProfessionId
}

type GetAllProfessionsJob struct {
	Err                error
	HrefReference      HrefReference
	ProfessionResponse ProfessionResponse
}

func (job GetAllProfessionsJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"href":  job.HrefReference.Href,
	}
}

func GetAllProfessions(opts GetAllProfessionsOptions) (chan GetAllProfessionsJob, error) {
	// querying index
	uri, err := opts.GetProfessionIndexURL()
	if err != nil {
		return nil, err
	}

	logging.WithField("opts.Blacklist", opts.Blacklist).Info("GetAllProfessions()")

	pIndex, _, err := NewProfessionsIndexResponseFromHTTP(uri)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to get profession-index")

		return nil, err
	}

	// starting up workers for gathering individual professions
	in := make(chan HrefReference)
	out := make(chan GetAllProfessionsJob)
	worker := func() {
		for hrefRef := range in {
			GetProfessionUri, err := opts.GetProfessionURL(hrefRef.Href)
			if err != nil {
				out <- GetAllProfessionsJob{
					Err:                err,
					HrefReference:      hrefRef,
					ProfessionResponse: ProfessionResponse{},
				}

				continue
			}

			profession, _, err := NewProfessionResponseFromHTTP(GetProfessionUri)
			if err != nil {
				out <- GetAllProfessionsJob{
					Err:                err,
					HrefReference:      hrefRef,
					ProfessionResponse: ProfessionResponse{},
				}

				continue
			}

			out <- GetAllProfessionsJob{
				Err:                nil,
				HrefReference:      hrefRef,
				ProfessionResponse: profession,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	// converting blacklist to map for filtering
	blacklistMap := map[ProfessionId]struct{}{}
	for _, id := range opts.Blacklist {
		blacklistMap[id] = struct{}{}
	}

	// queueing it up
	go func() {
		total := 0
		for _, profession := range pIndex.Professions {
			if _, ok := blacklistMap[profession.Id]; ok {
				logging.WithField("profession", profession.Id).Info("skipping profession fetching")

				continue
			}

			logging.WithField("profession", profession.Id).Info("fetching profession")

			in <- profession.Key

			total += 1
		}

		close(in)
	}()

	return out, nil
}
