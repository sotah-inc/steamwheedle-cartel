package state

import (
	"encoding/json"

	nats "github.com/nats-io/nats.go"
	AreaMapsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/areamaps"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
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

func NewAreaMapsState(mess messenger.Messenger, areaMapsDatabaseDir string) (AreaMapsState, error) {
	if err := util.EnsureDirExists(areaMapsDatabaseDir); err != nil {
		logging.WithField("error", err.Error()).Error("failed to ensure area-maps-database-dir exists")

		return AreaMapsState{}, err
	}

	areaMapsDatabase, err := AreaMapsDatabase.NewAreaMapsDatabase(areaMapsDatabaseDir)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise area-maps-database")

		return AreaMapsState{}, err
	}

	return AreaMapsState{
		Messenger:        mess,
		AreaMapsDatabase: areaMapsDatabase,
	}, nil
}

type AreaMapsState struct {
	Messenger        messenger.Messenger
	AreaMapsDatabase AreaMapsDatabase.Database
}

func (sta AreaMapsState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.AreaMaps:      sta.ListenForAreaMaps,
		subjects.AreaMapsQuery: sta.ListenForAreaMapsQuery,
	}
}

func (sta AreaMapsState) ListenForAreaMaps(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.AreaMaps), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		amRequest, err := NewAreaMapsRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		resolved, err := sta.AreaMapsDatabase.FindAreaMaps(amRequest.AreaMapIds)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		amRes := AreaMapsResponse{AreaMaps: resolved.SetUrls(gameversions.Retail)}
		data, err := amRes.EncodeForMessage()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = data
		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}

func (sta AreaMapsState) ListenForAreaMapsQuery(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.AreaMapsQuery), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		request, err := AreaMapsDatabase.NewQueryRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// querying the area-maps-database
		resp, err := sta.AreaMapsDatabase.AreaMapsQuery(request)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// marshalling for messenger
		encodedMessage, err := resp.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// dumping it out
		m.Data = string(encodedMessage)
		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
