package prod

import (
	"encoding/json"

	nats "github.com/nats-io/go-nats"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/messenger"
	mCodes "github.com/sotah-inc/steamwheedle-cartel/pkg/messenger/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
)

func (apiState ApiState) ListenForSessionSecret(stop state.ListenStopChan) error {
	err := apiState.IO.Messenger.Subscribe(string(subjects.SessionSecret), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		encodedData, err := json.Marshal(state.SessionSecretData{SessionSecret: apiState.SessionSecret.String()})
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			apiState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = string(encodedData)
		apiState.IO.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
