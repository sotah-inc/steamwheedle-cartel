package state

import (
	"encoding/json"

	"github.com/twinj/uuid"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	dCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/codes"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
)

// state
type State struct {
	RunID        uuid.UUID
	Listeners    Listeners
	BusListeners BusListeners
}

func DatabaseCodeToMessengerCode(dCode dCodes.Code) mCodes.Code {
	switch dCode {
	case dCodes.Ok:
		return mCodes.Ok
	case dCodes.Blank:
		return mCodes.Blank
	case dCodes.GenericError:
		return mCodes.GenericError
	case dCodes.MsgJSONParseError:
		return mCodes.MsgJSONParseError
	case dCodes.NotFound:
		return mCodes.NotFound
	case dCodes.UserError:
		return mCodes.UserError
	}

	return mCodes.Blank
}

func NewIntakeRequest(data []byte) (IntakeRequest, error) {
	out := IntakeRequest{}
	if err := json.Unmarshal(data, &out); err != nil {
		return IntakeRequest{}, err
	}

	return out, nil
}

type IntakeRequest struct {
	Tuples blizzardv2.RegionVersionConnectedRealmTuples `json:"tuples"`
}

func (req IntakeRequest) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(req)
}

func NewLoadIntakeRequest(data []byte) (LoadIntakeRequest, error) {
	out := LoadIntakeRequest{}
	if err := json.Unmarshal(data, &out); err != nil {
		return LoadIntakeRequest{}, err
	}

	return out, nil
}

type LoadIntakeRequest struct {
	Version gameversion.GameVersion             `json:"version"`
	Tuples  blizzardv2.LoadConnectedRealmTuples `json:"tuples"`
}

func (req LoadIntakeRequest) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(req)
}
