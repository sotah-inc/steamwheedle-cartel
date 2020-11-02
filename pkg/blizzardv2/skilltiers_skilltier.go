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

const skillTierURLFormat = "https://%s/data/wow/profession/%d/skill-tier/%d?namespace=static-%s"

func DefaultSkillTierURL(
	regionHostname string,
	professionId ProfessionId,
	skillTierId SkillTierId,
	regionName RegionName,
) string {
	return fmt.Sprintf(skillTierURLFormat, regionHostname, professionId, skillTierId, regionName)
}

type GetSkillTierURLFunc func(string) string

type SkillTierId int

type SkillTierCategoryRecipe struct {
	Key  HrefReference  `json:"key"`
	Name locale.Mapping `json:"name"`
	Id   RecipeId       `json:"recipe"`
}

type SkillTierCategory struct {
	Name    locale.Mapping            `json:"name"`
	Recipes []SkillTierCategoryRecipe `json:"recipes"`
}

type SkillTierResponse struct {
	LinksBase
	Id                SkillTierId         `json:"id"`
	Name              locale.Mapping      `json:"name"`
	MinimumSkillLevel int                 `json:"minimum_skill_level"`
	MaximumSkillLevel int                 `json:"maximum_skill_level"`
	Categories        []SkillTierCategory `json:"categories"`
}

func NewSkillTierResponse(body []byte) (SkillTierResponse, error) {
	psTier := &SkillTierResponse{}
	if err := json.Unmarshal(body, psTier); err != nil {
		return SkillTierResponse{}, err
	}

	return *psTier, nil
}

func NewSkillTierResponseFromHTTP(uri string) (SkillTierResponse, ResponseMeta, error) {
	resp, err := Download(DownloadOptions{Uri: uri})
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to download skill-tier")

		return SkillTierResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"status": resp.Status,
			"uri":    ClearAccessToken(uri),
		}).Error("resp from skill-tier was not 200")

		return SkillTierResponse{}, resp, errors.New("status was not 200")
	}

	psTier, err := NewSkillTierResponse(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to parse skill-tier response")

		return SkillTierResponse{}, resp, err
	}

	return psTier, resp, nil
}
