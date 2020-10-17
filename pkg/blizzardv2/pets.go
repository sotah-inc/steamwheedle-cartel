package blizzardv2

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetAllPetsOptions struct {
	GetPetIndexURL func() (string, error)
	GetPetURL      func(string) (string, error)

	Blacklist []PetId
	Limit     int
}

type GetAllPetsJob struct {
	Err           error
	HrefReference HrefReference
	PetResponse   PetResponse
}

func (job GetAllPetsJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"href":  job.HrefReference.Href,
	}
}

func GetAllPets(opts GetAllPetsOptions) (chan GetAllPetsJob, error) {
	// querying index
	uri, err := opts.GetPetIndexURL()
	if err != nil {
		return nil, err
	}

	pIndex, _, err := NewPetIndexFromHTTP(uri)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to get pet-index")

		return nil, err
	}

	// starting up workers for gathering individual pets
	in := make(chan HrefReference)
	out := make(chan GetAllPetsJob)
	worker := func() {
		for hrefRef := range in {
			getPetUri, err := opts.GetPetURL(hrefRef.Href)
			if err != nil {
				out <- GetAllPetsJob{
					Err:           err,
					HrefReference: hrefRef,
					PetResponse:   PetResponse{},
				}

				continue
			}

			cRealm, _, err := NewPetFromHTTP(getPetUri)
			if err != nil {
				out <- GetAllPetsJob{
					Err:           err,
					HrefReference: hrefRef,
					PetResponse:   PetResponse{},
				}

				continue
			}

			out <- GetAllPetsJob{
				Err:           nil,
				HrefReference: hrefRef,
				PetResponse:   cRealm,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	// converting blacklist to map for filtering
	blacklistMap := map[PetId]struct{}{}
	for _, id := range opts.Blacklist {
		blacklistMap[id] = struct{}{}
	}

	// queueing it up
	go func() {
		total := 0
		for _, pet := range pIndex.Pets {
			if total > opts.Limit {
				break
			}

			if _, ok := blacklistMap[pet.Id]; ok {
				continue
			}

			in <- pet.Key

			total += 1
		}

		close(in)
	}()

	return out, nil
}
