package state

import (
	nats "github.com/nats-io/go-nats"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta State) ListenForGenericTestErrors(stop ListenStopChan) error {
	err := sta.IO.Messenger.Subscribe(string(subjects.GenericTestErrors), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()
		m.Err = "Test error"
		m.Code = codes.GenericError
		sta.IO.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
