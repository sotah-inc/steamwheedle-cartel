package state

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	ProfessionsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/professions" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta ProfessionsState) ListenForProfessionsIntake(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.ProfessionsIntake), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		if err := sta.ProfessionsIntake(); err != nil {
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

func (sta ProfessionsState) ProfessionsIntake() error {
	startTime := time.Now()

	// resolving profession-ids to not sync
	professionIds, err := sta.ProfessionsDatabase.GetProfessionIds()
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to get profession-ids for blacklist")

		return err
	}

	logging.WithField(
		"professions-blacklist",
		len(professionIds),
	).Info("collecting professions sans blacklist")

	// starting up an intake queue
	getEncodedProfessionsOut, err := sta.LakeClient.GetEncodedProfessions(professionIds)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initiate encoded-professions fetching")

		return err
	}
	persistProfessionsIn := make(chan ProfessionsDatabase.PersistEncodedProfessionsInJob)

	// queueing it all up
	go func() {
		for job := range getEncodedProfessionsOut {
			if job.Err() != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to resolve profession")

				continue
			}

			logging.WithField("profession-id", job.Id()).Info("enqueueing profession for persistence")

			persistProfessionsIn <- ProfessionsDatabase.PersistEncodedProfessionsInJob{
				ProfessionId:      job.Id(),
				EncodedProfession: job.EncodedProfession(),
			}
		}

		close(persistProfessionsIn)
	}()

	totalPersisted, err := sta.ProfessionsDatabase.PersistEncodedProfessions(persistProfessionsIn)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to persist professions")

		return err
	}

	logging.WithFields(logrus.Fields{
		"total":          totalPersisted,
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total persisted in professions-intake")

	return nil
}
