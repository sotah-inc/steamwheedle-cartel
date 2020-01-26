package dev

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"

	nats "github.com/nats-io/go-nats"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/sortdirections"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/sortkinds"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func newAuctionsRequest(payload []byte) (AuctionsRequest, error) {
	ar := &AuctionsRequest{}
	err := json.Unmarshal(payload, &ar)
	if err != nil {
		return AuctionsRequest{}, err
	}

	return *ar, nil
}

type AuctionsRequest struct {
	RegionName    blizzard.RegionName          `json:"region_name"`
	RealmSlug     blizzard.RealmSlug           `json:"realm_slug"`
	Page          int                          `json:"page"`
	Count         int                          `json:"count"`
	SortDirection sortdirections.SortDirection `json:"sort_direction"`
	SortKind      sortkinds.SortKind           `json:"sort_kind"`
	ItemFilters   []blizzard.ItemID            `json:"item_filters"`
}

func (ar AuctionsRequest) resolve(laState LiveAuctionsState) (sotah.MiniAuctionList, state.RequestError) {
	realmLadbase, err := laState.IO.Databases.LiveAuctionsDatabases.GetDatabase(
		ar.RegionName,
		ar.RealmSlug,
	)
	if err != nil {
		return sotah.MiniAuctionList{}, state.RequestError{Code: codes.NotFound, Message: err.Error()}
	}

	if ar.Page < 0 {
		return sotah.MiniAuctionList{}, state.RequestError{Code: codes.UserError, Message: "Page must be >=0"}
	}
	if ar.Count == 0 {
		return sotah.MiniAuctionList{}, state.RequestError{Code: codes.UserError, Message: "Count must be >0"}
	} else if ar.Count > 1000 {
		return sotah.MiniAuctionList{}, state.RequestError{Code: codes.UserError, Message: "Count must be <=1000"}
	}

	maList, err := realmLadbase.GetMiniAuctionList()
	if err != nil {
		return sotah.MiniAuctionList{}, state.RequestError{Code: codes.GenericError, Message: err.Error()}
	}

	return maList, state.RequestError{Code: codes.Ok, Message: ""}
}

type auctionsResponse struct {
	AuctionList sotah.MiniAuctionList `json:"auctions"`
	Total       int                   `json:"total"`
	TotalCount  int                   `json:"total_count"`
}

func (ar auctionsResponse) encodeForMessage() (string, error) {
	jsonEncodedAuctions, err := json.Marshal(ar)
	if err != nil {
		return "", err
	}

	gzipEncodedAuctions, err := util.GzipEncode(jsonEncodedAuctions)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncodedAuctions), nil
}

func (laState LiveAuctionsState) ListenForAuctions(stop state.ListenStopChan) error {
	err := laState.IO.Messenger.Subscribe(string(subjects.Auctions), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		aRequest, err := newAuctionsRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// resolving data from State
		realmAuctions, reErr := aRequest.resolve(laState)
		if reErr.Code != codes.Ok {
			m.Err = reErr.Message
			m.Code = reErr.Code
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// initial response format
		aResponse := auctionsResponse{Total: -1, TotalCount: -1, AuctionList: realmAuctions}

		// filtering in auctions by items
		if len(aRequest.ItemFilters) > 0 {
			aResponse.AuctionList = aResponse.AuctionList.FilterByItemIDs(aRequest.ItemFilters)
		}

		// calculating the total for paging
		aResponse.Total = len(aResponse.AuctionList)

		// calculating the total-count for review
		totalCount := 0
		for _, mAuction := range realmAuctions {
			totalCount += len(mAuction.AucList)
		}
		aResponse.TotalCount = totalCount

		// optionally sorting
		if aRequest.SortKind != sortkinds.None && aRequest.SortDirection != sortdirections.None {
			err = aResponse.AuctionList.Sort(aRequest.SortKind, aRequest.SortDirection)
			if err != nil {
				m.Err = err.Error()
				m.Code = codes.UserError
				laState.IO.Messenger.ReplyTo(natsMsg, m)

				return
			}
		}

		// truncating the list
		aResponse.AuctionList, err = aResponse.AuctionList.Limit(aRequest.Count, aRequest.Page)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.UserError
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// encoding the auctions list for output
		data, err := aResponse.encodeForMessage()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = data
		laState.IO.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}

func (laState LiveAuctionsState) NewMiniAuctionsList(req AuctionsRequest) (sotah.MiniAuctionList, error) {
	encodedMessage, err := json.Marshal(req)
	if err != nil {
		return sotah.MiniAuctionList{}, err
	}

	msg, err := laState.IO.Messenger.Request(string(subjects.Auctions), encodedMessage)
	if err != nil {
		return sotah.MiniAuctionList{}, err
	}

	if msg.Code != codes.Ok {
		return sotah.MiniAuctionList{}, errors.New(msg.Err)
	}

	return sotah.NewMiniAuctionListFromGzipped([]byte(msg.Data))
}
