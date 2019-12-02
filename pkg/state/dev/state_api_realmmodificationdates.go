package dev

import (
	nats "github.com/nats-io/go-nats"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta APIState) ListenForRealmModificationDates(stop state.ListenStopChan) error {
	err := sta.IO.Messenger.Subscribe(string(subjects.RealmModificationDates), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		encodedData, err := sta.RegionRealmModificationDates.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = string(encodedData)
		sta.IO.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
