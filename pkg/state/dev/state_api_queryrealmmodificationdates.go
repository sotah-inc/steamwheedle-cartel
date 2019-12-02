package dev

import (
	"errors"

	"git.sotah.info/steamwheedle-cartel/pkg/blizzard"
	"git.sotah.info/steamwheedle-cartel/pkg/messenger"
	mCodes "git.sotah.info/steamwheedle-cartel/pkg/messenger/codes"
	"git.sotah.info/steamwheedle-cartel/pkg/sotah"
	"git.sotah.info/steamwheedle-cartel/pkg/state"
	"git.sotah.info/steamwheedle-cartel/pkg/state/subjects"
	nats "github.com/nats-io/go-nats"
)

func (sta APIState) ListenForQueryRealmModificationDates(stop state.ListenStopChan) error {
	err := sta.IO.Messenger.Subscribe(string(subjects.QueryRealmModificationDates), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		req, err := state.NewRealmModificationDatesRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		regionName, err := func() (blizzard.RegionName, error) {
			for _, region := range sta.Regions {
				if region.Name != blizzard.RegionName(req.RegionName) {
					continue
				}

				if _, ok := sta.Statuses[blizzard.RegionName(req.RegionName)]; !ok {
					continue
				}

				return region.Name, nil
			}

			return blizzard.RegionName(""), errors.New("region not found")
		}()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.NotFound
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		realm, err := func() (sotah.Realm, error) {
			for _, realm := range sta.Statuses[regionName].Realms {
				if realm.Slug != blizzard.RealmSlug(req.RealmSlug) {
					continue
				}

				return realm, nil
			}

			return sotah.Realm{}, errors.New("realm not found")
		}()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.NotFound
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		res := state.RealmModificationDatesResponse{
			RealmModificationDates: sta.RegionRealmModificationDates.Get(realm.Region.Name, realm.Slug),
		}

		encodedData, err := res.EncodeForDelivery()
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
