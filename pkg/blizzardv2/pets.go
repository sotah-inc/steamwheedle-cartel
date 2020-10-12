package blizzardv2

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetAllPetsOptions struct {
	GetPetIndexURL func() (string, error)
	GetPetURL      func(string) (string, error)
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

func GetAllPets(opts GetAllPetsOptions) ([]PetResponse, error) {
	// querying index
	uri, err := opts.GetPetIndexURL()
	if err != nil {
		return []PetResponse{}, err
	}

	pIndex, _, err := NewPetIndexFromHTTP(uri)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to get pet-index")

		return []PetResponse{}, err
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

	// queueing it up
	go func() {
		for _, pet := range pIndex.Pets {
			in <- pet.Key
		}

		close(in)
	}()

	// waiting for it all to drain out
	result := make([]PetResponse, len(pIndex.Pets))
	i := 0
	for outJob := range out {
		if outJob.Err != nil {
			return []PetResponse{}, outJob.Err
		}

		result[i] = outJob.PetResponse
		i += 1
	}

	return result, nil
}
