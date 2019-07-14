package prod

import (
	"encoding/json"

	nats "github.com/nats-io/go-nats"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/messenger"
	mCodes "github.com/sotah-inc/steamwheedle-cartel/pkg/messenger/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah/gameversions"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
)

func (apiState ApiState) ListenForReceiveRealms(stop state.ListenStopChan) error {
	err := apiState.IO.Messenger.Subscribe(string(subjects.ReceiveRealms), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		var regionRealmSlugs map[blizzard.RegionName][]blizzard.RealmSlug
		if err := json.Unmarshal(natsMsg.Data, &regionRealmSlugs); err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			apiState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		hellRegionRealms, err := apiState.IO.HellClient.GetRegionRealms(regionRealmSlugs, gameversions.Retail)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			apiState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		apiState.HellRegionRealms = apiState.HellRegionRealms.Merge(hellRegionRealms)

		apiState.IO.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
