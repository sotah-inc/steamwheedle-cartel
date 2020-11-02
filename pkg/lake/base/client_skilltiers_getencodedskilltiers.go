package base

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type GetEncodedSkillTierJob interface {
	Err() error
	Id() blizzardv2.SkillTierId
	EncodedSkillTier() []byte
	ToLogrusFields() logrus.Fields
}
