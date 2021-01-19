package state

import (
	PetsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/pets" // nolint:lll
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type NewPetsStateOptions struct {
	LakeClient BaseLake.Client
	Messenger  messenger.Messenger

	PetsDatabaseDir string
}

func NewPetsState(opts NewPetsStateOptions) (PetsState, error) {
	if err := util.EnsureDirExists(opts.PetsDatabaseDir); err != nil {
		logging.WithField("error", err.Error()).Error("failed to ensure pets-database-dir exists")

		return PetsState{}, err
	}

	petsDatabase, err := PetsDatabase.NewDatabase(opts.PetsDatabaseDir)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise pets-database")

		return PetsState{}, err
	}

	return PetsState{
		LakeClient:   opts.LakeClient,
		Messenger:    opts.Messenger,
		PetsDatabase: petsDatabase,
	}, nil
}

type PetsState struct {
	LakeClient   BaseLake.Client
	Messenger    messenger.Messenger
	PetsDatabase PetsDatabase.Database
}

func (sta PetsState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.Pets:       sta.ListenForPets,
		subjects.PetsIntake: sta.ListenForPetsIntake,
		subjects.PetsQuery:  sta.ListenForPetsQuery,
	}
}
