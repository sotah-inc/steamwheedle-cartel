package codes

import "net/http"

// Code - typehint for these enums
type Code int

/*
Codes - message response codes
*/
var (
	NoAction          Code = 2
	Ok                Code = 1
	Blank             Code // zero value is 0
	GenericError      Code = -1
	MsgJSONParseError Code = -2
	NotFound          Code = -3
	UserError         Code = -4
	BlizzardError     Code = -5
)

func CodeToHTTPStatus(code Code) int {
	switch code {
	case NoAction:
		return http.StatusNotModified
	case Ok:
		return http.StatusOK
	case NotFound:
		return http.StatusNotFound
	case UserError:
		return http.StatusBadRequest
	case Blank:
	case GenericError:
	case MsgJSONParseError:
	case BlizzardError:
	default:
		return http.StatusInternalServerError
	}

	return http.StatusInternalServerError
}
