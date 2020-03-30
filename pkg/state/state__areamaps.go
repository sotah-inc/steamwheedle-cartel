package state

import (
	"encoding/json"

	nats "github.com/nats-io/go-nats"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func NewAreaMapsRequest(payload []byte) (AreaMapsRequest, error) {
	amRequest := &AreaMapsRequest{}
	err := json.Unmarshal(payload, &amRequest)
	if err != nil {
		return AreaMapsRequest{}, err
	}

	return *amRequest, nil
}

type AreaMapsRequest struct {
	AreaMapIds []sotah.AreaMapId `json:"areaMapIds"`
}

type AreaMapsResponse struct {
	AreaMaps sotah.AreaMapMap `json:"areaMaps"`
}

func (amRes AreaMapsResponse) EncodeForMessage() (string, error) {
	result, err := json.Marshal(amRes)
	if err != nil {
		return "", err
	}

	return string(result), err
}

type AreaMapsState struct {
	messenger        messenger.Messenger
	areaMapsDatabase database.AreaMapsDatabase
}

func (sta AreaMapsState) ListenForAreaMaps(stop ListenStopChan) error {
	err := sta.messenger.Subscribe(string(subjects.AreaMaps), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		amRequest, err := NewAreaMapsRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.messenger.ReplyTo(natsMsg, m)

			return
		}

		resolved, err := sta.areaMapsDatabase.FindAreaMaps(amRequest.AreaMapIds)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.messenger.ReplyTo(natsMsg, m)

			return
		}

		amRes := AreaMapsResponse{AreaMaps: resolved.SetUrls(gameversions.Retail)}
		data, err := amRes.EncodeForMessage()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = data
		sta.messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}

func (sta AreaMapsState) ListenForAreaMapsQuery(stop ListenStopChan) error {
	err := sta.messenger.Subscribe(string(subjects.AreaMapsQuery), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		request, err := database.NewAreaMapsQueryRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.MsgJSONParseError
			sta.messenger.ReplyTo(natsMsg, m)

			return
		}

		// querying the area-maps-database
		resp, err := sta.areaMapsDatabase.AreaMapsQuery(request)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.messenger.ReplyTo(natsMsg, m)

			return
		}

		// marshalling for messenger
		encodedMessage, err := resp.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.messenger.ReplyTo(natsMsg, m)

			return
		}

		// dumping it out
		m.Data = string(encodedMessage)
		sta.messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
