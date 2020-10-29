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

type getEncodedProfessionJob struct {
	err                   error
	id                    blizzardv2.ProfessionId
	encodedProfession     []byte
	encodedNormalizedName []byte
}

func (g getEncodedProfessionJob) Err() error                    { return g.err }
func (g getEncodedProfessionJob) Id() blizzardv2.ProfessionId   { return g.id }
func (g getEncodedProfessionJob) EncodedProfession() []byte     { return g.encodedProfession }
func (g getEncodedProfessionJob) EncodedNormalizedName() []byte { return g.encodedNormalizedName }
func (g getEncodedProfessionJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": g.err.Error(),
		"id":    g.id,
	}
}

func (client Client) GetEncodedProfessions(
	blacklist []blizzardv2.ProfessionId,
) (chan BaseLake.GetEncodedProfessionJob, error) {
	out := make(chan BaseLake.GetEncodedProfessionJob)

	// starting up workers for resolving items
	professionsOut, err := client.resolveProfessions(blacklist)
	if err != nil {
		return nil, err
	}

	// queueing it all up
	go func() {
		for job := range professionsOut {
			if job.Err != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to resolve profession")

				continue
			}

			normalizedName, err := func() (locale.Mapping, error) {
				foundName, ok := job.ProfessionResponse.Name[locale.EnUS]
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
					"response": job.ProfessionResponse,
				}).Error("failed to normalize name")

				continue
			}

			profession := sotah.Profession{
				BlizzardMeta: job.ProfessionResponse,
				SotahMeta:    sotah.ProfessionMeta{},
			}

			encodedProfession, err := profession.EncodeForStorage()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":      err.Error(),
					"profession": profession.BlizzardMeta.Id,
				}).Error("failed to encode profession for storage")

				continue
			}

			encodedNormalizedName, err := normalizedName.EncodeForStorage()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error": err.Error(),
					"item":  profession.BlizzardMeta.Id,
				}).Error("failed to encode normalized-name for storage")

				continue
			}

			out <- getEncodedProfessionJob{
				id:                    job.ProfessionResponse.Id,
				encodedProfession:     encodedProfession,
				encodedNormalizedName: encodedNormalizedName,
			}
		}

		close(out)
	}()

	return out, nil
}
