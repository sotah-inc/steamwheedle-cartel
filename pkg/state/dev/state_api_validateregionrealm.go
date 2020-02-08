package dev

import (
	nats "github.com/nats-io/go-nats"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta *APIState) ListenForValidateRegionRealm(stop state.ListenStopChan) error {
	err := sta.IO.Messenger.Subscribe(string(subjects.ValidateRegionRealm), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		request, err := state.NewValidateRegionRealmRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.MsgJSONParseError
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		res := state.ValidateRegionRealmResponse{IsValid: false}

		hasRegion := func() bool {
			for _, reg := range sta.Regions {
				if string(reg.Name) == request.RegionName {
					return true
				}
			}

			return false
		}()
		if !hasRegion {
			encodedMessage, err := res.EncodeForMessage()
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.MsgJSONParseError
				sta.IO.Messenger.ReplyTo(natsMsg, m)

				return
			}

			m.Data = encodedMessage
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		hasRealm := func() bool {
			status, ok := sta.Statuses[blizzard.RegionName(request.RegionName)]
			if !ok {
				return false
			}

			_, ok = status.Realms.ToRealmMap()[blizzard.RealmSlug(request.RealmSlug)]

			return ok
		}()
		if !hasRealm {
			encodedMessage, err := res.EncodeForMessage()
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.MsgJSONParseError
				sta.IO.Messenger.ReplyTo(natsMsg, m)

				return
			}

			m.Data = encodedMessage
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		res.IsValid = true

		// marshalling for messenger
		encodedMessage, err := res.EncodeForMessage()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// dumping it out
		m.Data = string(encodedMessage)
		sta.IO.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
