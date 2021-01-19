package state

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	PetsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/pets" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta PetsState) ListenForPetsIntake(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.PetsIntake), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		if err := sta.petsIntake(); err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}

func (sta PetsState) petsIntake() error {
	startTime := time.Now()

	// resolving pet-ids to not sync
	petIds, err := sta.PetsDatabase.GetPetIds()
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to get pet-ids for blacklist")

		return err
	}

	logging.WithField("pets-blacklist", len(petIds)).Info("collecting pets sans blacklist")

	// starting up an intake queue
	getEncodedPetsOut, err := sta.LakeClient.GetEncodedPets(petIds)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initiate encoded-pets fetching")

		return err
	}
	persistPetsIn := make(chan PetsDatabase.PersistEncodedPetsInJob)

	// queueing it all up
	go func() {
		for job := range getEncodedPetsOut {
			if job.Err() != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to resolve pet")

				continue
			}

			logging.WithField("pet-id", job.Id()).Info("enqueueing pet for persistence")

			persistPetsIn <- PetsDatabase.PersistEncodedPetsInJob{
				Id:                    job.Id(),
				EncodedPet:            job.EncodedPet(),
				EncodedNormalizedName: job.EncodedNormalizedName(),
			}
		}

		close(persistPetsIn)
	}()

	totalPersisted, err := sta.PetsDatabase.PersistEncodedPets(persistPetsIn)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to persist pets")

		return err
	}

	logging.WithFields(logrus.Fields{
		"total":          totalPersisted,
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total persisted in pets-intake")

	return nil
}
