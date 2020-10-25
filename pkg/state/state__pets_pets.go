package state

import (
	"encoding/base64"
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"

	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewPetsRequest(payload []byte) (PetsRequest, error) {
	pRequest := &PetsRequest{}
	err := json.Unmarshal(payload, &pRequest)
	if err != nil {
		return PetsRequest{}, err
	}

	return *pRequest, nil
}

type PetsRequest struct {
	Locale locale.Locale      `json:"locale"`
	PetIds []blizzardv2.PetId `json:"petIds"`
}

type PetsResponse struct {
	Pets []sotah.ShortPet `json:"pets"`
}

func (pResponse PetsResponse) EncodeForMessage() (string, error) {
	encodedResult, err := json.Marshal(pResponse)
	if err != nil {
		return "", err
	}

	gzippedResult, err := util.GzipEncode(encodedResult)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzippedResult), nil
}

func (sta PetsState) ListenForPets(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.Pets), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		pRequest, err := NewPetsRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		if pRequest.Locale.IsZero() {
			m.Err = "locale was zero"
			m.Code = codes.UserError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		foundPets, err := sta.PetsDatabase.FindPets(pRequest.PetIds)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		resolvedShortPets := sotah.NewShortPetList(foundPets, pRequest.Locale)

		pResponse := PetsResponse{Pets: resolvedShortPets}
		data, err := pResponse.EncodeForMessage()
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
