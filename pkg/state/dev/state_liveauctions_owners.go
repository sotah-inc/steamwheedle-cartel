package dev

import (
	"encoding/json"
	"errors"
	"sort"

	nats "github.com/nats-io/go-nats"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/messenger"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/messenger/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
)

func newOwnersRequest(payload []byte) (OwnersRequest, error) {
	request := &OwnersRequest{}
	err := json.Unmarshal(payload, &request)
	if err != nil {
		return OwnersRequest{}, err
	}

	return *request, nil
}

type OwnersRequest struct {
	RegionName blizzard.RegionName `json:"region_name"`
	RealmSlug  blizzard.RealmSlug  `json:"realm_slug"`
	Query      string              `json:"query"`
}

func (request OwnersRequest) resolve(laState LiveAuctionsState) (sotah.MiniAuctionList, error) {
	regionLadBases, ok := laState.IO.Databases.LiveAuctionsDatabases[request.RegionName]
	if !ok {
		return sotah.MiniAuctionList{}, errors.New("invalid region name")
	}

	ladBase, ok := regionLadBases[request.RealmSlug]
	if !ok {
		return sotah.MiniAuctionList{}, errors.New("invalid realm slug")
	}

	maList, err := ladBase.GetMiniAuctionList()
	if err != nil {
		return sotah.MiniAuctionList{}, err
	}

	return maList, nil
}

func (laState LiveAuctionsState) ListenForOwners(stop state.ListenStopChan) error {
	err := laState.IO.Messenger.Subscribe(string(subjects.Owners), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		request, err := newOwnersRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// resolving mini-auctions-list from the request and State
		mal, err := request.resolve(laState)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.NotFound
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		o, err := sotah.NewOwnersFromAuctions(mal)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// optionally filtering in matches
		if request.Query != "" {
			o.Owners = o.Owners.Filter(request.Query)
		}

		// sorting and truncating
		sort.Sort(sotah.OwnersByName(o.Owners))
		o.Owners = o.Owners.Limit()

		// marshalling for Messenger
		encodedMessage, err := json.Marshal(o)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// dumping it out
		m.Data = string(encodedMessage)
		laState.IO.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
