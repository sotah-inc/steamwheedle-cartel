package disk

import (
	"errors"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type getEncodedPetJob struct {
	err                   error
	id                    blizzardv2.PetId
	encodedPet            []byte
	encodedNormalizedName []byte
}

func (g getEncodedPetJob) Err() error                    { return g.err }
func (g getEncodedPetJob) Id() blizzardv2.PetId          { return g.id }
func (g getEncodedPetJob) EncodedPet() []byte            { return g.encodedPet }
func (g getEncodedPetJob) EncodedNormalizedName() []byte { return g.encodedNormalizedName }
func (g getEncodedPetJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": g.err.Error(),
		"id":    g.id,
	}
}

func (client Client) GetEncodedPets(blacklist []blizzardv2.PetId) chan BaseLake.GetEncodedPetJob {
	out := make(chan BaseLake.GetEncodedPetJob)

	// starting up workers for resolving items
	petsOut := client.resolvePets(blacklist)

	// queueing it all up
	go func() {
		for job := range petsOut {
			if job.Err != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to resolve pet")

				continue
			}

			normalizedName, err := func() (locale.Mapping, error) {
				foundName, ok := job.PetResponse.Name[locale.EnUS]
				if !ok {
					return locale.Mapping{}, errors.New("failed to resolve enUS name")
				}

				normalizedName, err := sotah.NormalizeString(foundName)
				if err != nil {
					return locale.Mapping{}, err
				}

				return locale.Mapping{locale.EnUS: normalizedName}, nil
			}()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":    err.Error(),
					"response": job.PetResponse,
				}).Error("failed to normalize name")

				continue
			}

			pet := sotah.Pet{
				BlizzardMeta: job.PetResponse,
				SotahMeta: sotah.PetMeta{
					NormalizedName: normalizedName,
				},
			}

			encodedPet, err := pet.EncodeForStorage()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error": err.Error(),
					"pet":   pet.BlizzardMeta.Id,
				}).Error("failed to encode pet for storage")

				continue
			}

			encodedNormalizedName, err := normalizedName.EncodeForStorage()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error": err.Error(),
					"item":  pet.BlizzardMeta.Id,
				}).Error("failed to encode normalized-name for storage")

				continue
			}

			out <- getEncodedPetJob{
				id:                    job.PetResponse.Id,
				encodedPet:            encodedPet,
				encodedNormalizedName: encodedNormalizedName,
			}
		}

		close(out)
	}()

	return out
}
