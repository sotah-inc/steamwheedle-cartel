package prod

import (
	nats "github.com/nats-io/go-nats"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/messenger"
	mCodes "github.com/sotah-inc/steamwheedle-cartel/pkg/messenger/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
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
