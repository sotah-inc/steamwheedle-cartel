package state

import (
	"encoding/base64"
	"encoding/json"

	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewReceiveRegionTimestampsRequest(
	base64Encoded string,
) (ReceiveRegionTimestampsRequest, error) {
	gzipEncoded, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return ReceiveRegionTimestampsRequest{}, err
	}

	jsonEncoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return ReceiveRegionTimestampsRequest{}, err
	}

	out := ReceiveRegionTimestampsRequest{}
	if err := json.Unmarshal(jsonEncoded, &out); err != nil {
		return ReceiveRegionTimestampsRequest{}, err
	}

	return out, nil
}

type ReceiveRegionTimestampsRequest struct {
	Version          gameversion.GameVersion `json:"game_version"`
	RegionTimestamps sotah.RegionTimestamps  `json:"region_timestamps"`
}

func (sta RegionsState) ListenForReceiveRegionTimestamps(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(
		string(subjects.ReceiveRegionTimestamps),
		stop,
		func(natsMsg nats.Msg) {
			m := messenger.NewMessage()

			req, err := NewReceiveRegionTimestampsRequest(m.Data)
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			if !sta.GameVersionList.Includes(req.Version) {
				m.Err = "invalid game-version"
				m.Code = mCodes.UserError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			if err := sta.RegionsDatabase.ReceiveRegionTimestamps(
				req.Version,
				req.RegionTimestamps,
			); err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			sta.Messenger.ReplyTo(natsMsg, m)
		},
	)
	if err != nil {
		return err
	}

	return nil
}
