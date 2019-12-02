package sotah

import "git.sotah.info/steamwheedle-cartel/pkg/sotah/codes"

func NewErrorMessage(err error) Message {
	out := NewMessage()
	out.Code = codes.GenericError
	out.Err = err.Error()

	return out
}

func NewMessage() Message {
	return Message{Code: codes.Ok}
}

type Message struct {
	Data string     `json:"data"`
	Err  string     `json:"error"`
	Code codes.Code `json:"code"`
}
