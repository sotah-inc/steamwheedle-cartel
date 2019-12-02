package prod

import (
	"git.sotah.info/steamwheedle-cartel/pkg/messenger"
	mCodes "git.sotah.info/steamwheedle-cartel/pkg/messenger/codes"
	"git.sotah.info/steamwheedle-cartel/pkg/state"
	"git.sotah.info/steamwheedle-cartel/pkg/state/subjects"
	nats "github.com/nats-io/go-nats"
)

func (apiState ApiState) ListenForRealmModificationDates(stop state.ListenStopChan) error {
	err := apiState.IO.Messenger.Subscribe(string(subjects.RealmModificationDates), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		encodedData, err := apiState.HellRegionRealms.ToRegionRealmModificationDates().EncodeForDelivery()
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
