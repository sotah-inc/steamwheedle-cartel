package state

import (
	"github.com/twinj/uuid"
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
