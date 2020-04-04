package state

import (
	"encoding/json"
	"fmt"

	nats "github.com/nats-io/go-nats"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type NewRegionStateOptions struct {
	BlizzardState BlizzardState
	Regions       sotah.RegionList
	Messenger     messenger.Messenger
}

func NewRegionState(opts NewRegionStateOptions) (*RegionsState, error) {
	regionConnectedRealms, err := opts.BlizzardState.ResolveRegionConnectedRealms(opts.Regions)
	if err != nil {
		return nil, err
	}

	regionComposites := make(sotah.RegionComposites, len(opts.Regions))
	for i, region := range opts.Regions {
		connectedRealms := regionConnectedRealms[region.Name]

		realmComposites := make([]sotah.RealmComposite, len(connectedRealms))
		for j, response := range connectedRealms {
			realmComposites[j] = sotah.RealmComposite{ConnectedRealmResponse: response}
		}

		regionComposites[i] = sotah.RegionComposite{
			ConfigRegion:             region,
			ConnectedRealmComposites: realmComposites,
		}
	}

	return &RegionsState{
		BlizzardState:    opts.BlizzardState,
		Messenger:        opts.Messenger,
		RegionComposites: regionComposites,
	}, nil
}

type RegionsState struct {
	BlizzardState BlizzardState
	Messenger     messenger.Messenger

	RegionComposites sotah.RegionComposites
}

func (sta RegionsState) ReceiveTimestamps(timestamps sotah.RegionTimestamps) {
	sta.RegionComposites = sta.RegionComposites.Receive(timestamps)
}

func (sta RegionsState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.Status:              sta.ListenForStatus,
		subjects.ValidateRegionRealm: sta.ListenForValidateRegionRealm,
	}
}

func NewStatusRequest(payload []byte) (StatusRequest, error) {
	sr := &StatusRequest{}
	err := json.Unmarshal(payload, &sr)
	if err != nil {
		return StatusRequest{}, err
	}

	return *sr, nil
}

type StatusRequest struct {
	RegionName blizzardv2.RegionName `json:"region_name"`
}

func (sta RegionsState) ListenForStatus(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.Status), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		sRequest, err := NewStatusRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		region, err := sta.RegionComposites.FindBy(sRequest.RegionName)
		if err != nil {
			m.Err = fmt.Sprintf("invalid region name: %s", sRequest.RegionName)
			m.Code = codes.UserError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		encodedStatus, err := region.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = string(encodedStatus)
		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}

func NewValidateRegionRealmRequest(data []byte) (ValidateRegionRealmRequest, error) {
	var out ValidateRegionRealmRequest
	if err := json.Unmarshal(data, &out); err != nil {
		return ValidateRegionRealmRequest{}, err
	}

	return out, nil
}

type ValidateRegionRealmRequest struct {
	RegionName blizzardv2.RegionName `json:"region_name"`
	RealmSlug  blizzardv2.RealmSlug  `json:"realm_slug"`
}

type ValidateRegionRealmResponse struct {
	IsValid bool `json:"is_valid"`
}

func (res ValidateRegionRealmResponse) EncodeForDelivery() (string, error) {
	encodedResult, err := json.Marshal(res)
	if err != nil {
		return "", err
	}

	return string(encodedResult), nil
}

func (sta RegionsState) ListenForValidateRegionRealm(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.ValidateRegionRealm), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		req, err := NewValidateRegionRealmRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		res := ValidateRegionRealmResponse{
			IsValid: sta.RegionComposites.RegionRealmExists(req.RegionName, req.RealmSlug),
		}
		encoded, err := res.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = string(encoded)
		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
