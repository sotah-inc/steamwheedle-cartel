package state

import (
	ProfessionsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/professions"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type NewProfessionsStateOptions struct {
	LakeClient BaseLake.Client
	Messenger  messenger.Messenger

	ProfessionsDatabaseDir string
}

func NewProfessionsState(opts NewProfessionsStateOptions) (ProfessionsState, error) {
	if err := util.EnsureDirExists(opts.ProfessionsDatabaseDir); err != nil {
		logging.WithField("error", err.Error()).Error("failed to ensure professions-database-dir exists")

		return ProfessionsState{}, err
	}

	professionsDatabase, err := ProfessionsDatabase.NewDatabase(opts.ProfessionsDatabaseDir)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise professions-database")

		return ProfessionsState{}, err
	}

	return ProfessionsState{
		LakeClient:          opts.LakeClient,
		Messenger:           opts.Messenger,
		ProfessionsDatabase: professionsDatabase,
	}, nil
}

type ProfessionsState struct {
	LakeClient          BaseLake.Client
	Messenger           messenger.Messenger
	ProfessionsDatabase ProfessionsDatabase.Database
}

func (sta ProfessionsState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.ProfessionsIntake: sta.ListenForProfessionsIntake,
		subjects.SkillTiersIntake:  sta.ListenForSkillTiersIntake,
		subjects.Professions:       sta.ListenForProfessions,
		subjects.RecipesIntake:     sta.ListenForRecipesIntake,
		subjects.SkillTier:         sta.ListenForSkillTier,
		subjects.Recipe:            sta.ListenForRecipe,
	}
}
