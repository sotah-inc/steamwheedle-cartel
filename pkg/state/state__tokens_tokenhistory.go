package state

import (
	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type RegionTokenHistoryResponse struct {
	RegionName string `json:"region_name"`
}

func (sta TokensState) ListenForTokenHistory(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.TokenHistory), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// dumping it out
		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
