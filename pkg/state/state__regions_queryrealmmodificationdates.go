package state

import (
	"encoding/json"

	nats "github.com/nats-io/go-nats"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func NewRealmModificationDatesRequest(data []byte) (RealmModificationDatesRequest, error) {
	var r RealmModificationDatesRequest
	if err := json.Unmarshal(data, &r); err != nil {
		return RealmModificationDatesRequest{}, err
	}

	return r, nil
}

type RealmModificationDatesRequest struct {
	RegionName string `json:"region_name"`
	RealmSlug  string `json:"realm_slug"`
}

type RealmModificationDatesResponse struct {
	sotah.RealmModificationDates
}

func (r RealmModificationDatesResponse) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(r)
}

func (sta RegionsState) ListenForQueryRealmModificationDates(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.QueryRealmModificationDates), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		req, err := NewRealmModificationDatesRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		exists := sta.RegionComposites.RegionRealmExists(req.RegionName, req.RealmSlug)

		res := RealmModificationDatesResponse{
			RealmModificationDates: sta.RegionRealmModificationDates.Get(realm.Region.Name, realm.Slug),
		}

		encodedData, err := res.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = string(encodedData)
		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
