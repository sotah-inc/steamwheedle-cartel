package state

import (
	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta RegionsState) ListenForReceiveRegionTimestamps(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(
		string(subjects.ReceiveRegionTimestamps),
		stop,
		func(natsMsg nats.Msg) {
			m := messenger.NewMessage()

			regionTimestamps, err := sotah.NewRegionTimestamps(m.Data)
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			if err := sta.RegionsDatabase.ReceiveRegionTimestamps(regionTimestamps); err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			sta.Messenger.ReplyTo(natsMsg, m)
		},
	)
	if err != nil {
		return err
	}

	return nil
}
