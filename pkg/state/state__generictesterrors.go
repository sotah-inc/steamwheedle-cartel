package state

import (
	"git.sotah.info/steamwheedle-cartel/pkg/messenger"
	"git.sotah.info/steamwheedle-cartel/pkg/messenger/codes"
	"git.sotah.info/steamwheedle-cartel/pkg/state/subjects"
	nats "github.com/nats-io/go-nats"
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
