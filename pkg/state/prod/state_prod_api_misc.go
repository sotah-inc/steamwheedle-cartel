package prod

import (
	"encoding/json"

	"git.sotah.info/steamwheedle-cartel/pkg/state"

	"git.sotah.info/steamwheedle-cartel/pkg/messenger"
	mCodes "git.sotah.info/steamwheedle-cartel/pkg/messenger/codes"
	"git.sotah.info/steamwheedle-cartel/pkg/state/subjects"
	nats "github.com/nats-io/go-nats"
)

func (apiState ApiState) ListenForMessengerBoot(stop state.ListenStopChan) error {
	err := apiState.IO.Messenger.Subscribe(string(subjects.Boot), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		encodedResponse, err := json.Marshal(state.BootResponse{
			Regions:     apiState.Regions,
			ItemClasses: apiState.ItemClasses,
			Expansions:  apiState.Expansions,
			Professions: apiState.Professions,
		})
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.MsgJSONParseError
			apiState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = string(encodedResponse)
		apiState.IO.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
