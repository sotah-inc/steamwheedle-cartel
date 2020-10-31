package blizzardv2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

const professionSkillTierURLFormat = "https://%s/data/wow/profession/%d/skill-tier/%d?namespace=static-%s"

func DefaultProfessionSkillTierURL(
	regionHostname string,
	professionId ProfessionId,
	skillTierId ProfessionSkillTierId,
	regionName RegionName,
) string {
	return fmt.Sprintf(professionSkillTierURLFormat, regionHostname, professionId, skillTierId, regionName)
}

type GetProfessionSkillTierURLFunc func(string) string

type ProfessionSkillTierId int

type ProfessionSkillTierCategoryRecipe struct {
	Key  HrefReference  `json:"key"`
	Name locale.Mapping `json:"name"`
	Id   RecipeId       `json:"recipe"`
}

type ProfessionSkillTierCategory struct {
	Name    locale.Mapping                      `json:"name"`
	Recipes []ProfessionSkillTierCategoryRecipe `json:"recipes"`
}

type ProfessionSkillTierResponse struct {
	LinksBase
	Id                ProfessionSkillTierId         `json:"id"`
	Name              locale.Mapping                `json:"name"`
	MinimumSkillLevel int                           `json:"minimum_skill_level"`
	MaximumSkillLevel int                           `json:"maximum_skill_level"`
	Categories        []ProfessionSkillTierCategory `json:"categories"`
}

func NewProfessionSkillTierResponse(body []byte) (ProfessionSkillTierResponse, error) {
	psTier := &ProfessionSkillTierResponse{}
	if err := json.Unmarshal(body, psTier); err != nil {
		return ProfessionSkillTierResponse{}, err
	}

	return *psTier, nil
}

func NewProfessionSkillTierResponseFromHTTP(uri string) (ProfessionSkillTierResponse, ResponseMeta, error) {
	resp, err := Download(DownloadOptions{Uri: uri})
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to download profession skill-tier")

		return ProfessionSkillTierResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"status": resp.Status,
			"uri":    ClearAccessToken(uri),
		}).Error("resp from profession skill-tier was not 200")

		return ProfessionSkillTierResponse{}, resp, errors.New("status was not 200")
	}

	psTier, err := NewProfessionSkillTierResponse(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to parse profession skill-tier response")

		return ProfessionSkillTierResponse{}, resp, err
	}

	return psTier, resp, nil
}
