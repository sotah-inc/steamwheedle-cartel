package prod

import (
	nats "github.com/nats-io/go-nats"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (apiState ApiState) ListenForQueryRealmModificationDates(stop state.ListenStopChan) error {
	err := apiState.IO.Messenger.Subscribe(string(subjects.QueryRealmModificationDates), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		req, err := state.NewRealmModificationDatesRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			apiState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		logging.WithField("hell-region-realms", apiState.HellRegionRealms.Total()).Info("Checking hell-region-realms")

		hellRealms, ok := apiState.HellRegionRealms[blizzard.RegionName(req.RegionName)]
		if !ok {
			m.Err = "region not found"
			m.Code = mCodes.NotFound
			apiState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		hellRealm, ok := hellRealms[blizzard.RealmSlug(req.RealmSlug)]
		if !ok {
			m.Err = "realm not found"
			m.Code = mCodes.NotFound
			apiState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		res := state.RealmModificationDatesResponse{
			RealmModificationDates: sotah.RealmModificationDates{
				Downloaded:                 int64(hellRealm.Downloaded),
				LiveAuctionsReceived:       int64(hellRealm.LiveAuctionsReceived),
				PricelistHistoriesReceived: int64(hellRealm.PricelistHistoriesReceived),
			},
		}

		encodedData, err := res.EncodeForDelivery()
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
