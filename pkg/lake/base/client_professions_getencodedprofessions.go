package base

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type GetEncodedProfessionJob interface {
	Err() error
	Id() blizzardv2.ProfessionId
	EncodedProfession() []byte
	EncodedNormalizedName() []byte
	ToLogrusFields() logrus.Fields
}
