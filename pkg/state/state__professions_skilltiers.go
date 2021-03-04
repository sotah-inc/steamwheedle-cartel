package state

import (
	"encoding/base64"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewSkillTiersRequest(body []byte) (SkillTiersRequest, error) {
	out := SkillTiersRequest{}
	if err := json.Unmarshal(body, &out); err != nil {
		return SkillTiersRequest{}, err
	}

	return out, nil
}

type SkillTiersRequest struct {
	ProfessionId blizzardv2.ProfessionId  `json:"profession_id"`
	SkillTierIds []blizzardv2.SkillTierId `json:"skilltier_ids"`
	Locale       locale.Locale            `json:"locale"`
}

func (req SkillTiersRequest) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(req)
}

func NewSkillTiersResponse(base64Encoded string) (SkillTiersResponse, error) {
	gzipEncoded, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return SkillTiersResponse{}, err
	}

	gzipDecoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return SkillTiersResponse{}, err
	}

	out := SkillTiersResponse{}
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return SkillTiersResponse{}, err
	}

	return out, nil
}

type SkillTiersResponse struct {
	SkillTiers []sotah.ShortSkillTier `json:"skilltiers"`
}

func (resp SkillTiersResponse) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}

func (sta ProfessionsState) ListenForSkillTiers(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.SkillTiers), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		req, err := NewSkillTiersRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		skillTiers, err := sta.ProfessionsDatabase.GetSkillTiers(req.ProfessionId, req.SkillTierIds)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		respSkillTiers := make([]sotah.ShortSkillTier, len(skillTiers))
		for i, skillTier := range skillTiers {
			recipes, err := sta.ProfessionsDatabase.GetRecipes(skillTier.RecipeIds())
			if err != nil {
				m.Err = err.Error()
				m.Code = codes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			recipesMap := map[blizzardv2.RecipeId]sotah.Recipe{}
			for _, recipe := range recipes {
				recipesMap[recipe.BlizzardMeta.Id] = recipe
			}

			respSkillTiers[i] = sotah.NewShortSkillTier(skillTier, req.Locale, recipesMap)
		}

		resp := SkillTiersResponse{SkillTiers: respSkillTiers}
		data, err := resp.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
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
