package professions

import (
	"fmt"
	"strconv"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/professions/professionsflags" // nolint:lll
)

// bucketing
func baseBucketName() []byte {
	return []byte("professions")
}

func skillTiersBucketName(professionId blizzardv2.ProfessionId) []byte {
	return []byte(fmt.Sprintf("profession-%d-skill-tiers", professionId))
}

func recipesBucketName() []byte {
	return []byte("recipes")
}

func recipeNamesBucketName() []byte {
	return []byte("recipe-names")
}

func itemsCraftedByBucketName() []byte {
	return []byte("items-crafted-by")
}

// base keying
func baseKeyName(id blizzardv2.ProfessionId) []byte {
	return []byte(fmt.Sprintf("profession-%d", id))
}

func professionIdFromKeyName(key []byte) (blizzardv2.ProfessionId, error) {
	unparsedId, err := strconv.Atoi(string(key)[len("profession-"):])
	if err != nil {
		return blizzardv2.ProfessionId(0), err
	}

	return blizzardv2.ProfessionId(unparsedId), nil
}

func flagsBucketName() []byte {
	return []byte("flags")
}

// skill-tiers keying
func skillTiersKeyName(id blizzardv2.SkillTierId) []byte {
	return []byte(fmt.Sprintf("skill-tier-%d", id))
}

func skillTierIdFromKeyName(key []byte) (blizzardv2.SkillTierId, error) {
	unparsedId, err := strconv.Atoi(string(key)[len("skill-tier-"):])
	if err != nil {
		return blizzardv2.SkillTierId(0), err
	}

	return blizzardv2.SkillTierId(unparsedId), nil
}

// recipes keying
func recipeKeyName(id blizzardv2.RecipeId) []byte {
	return []byte(fmt.Sprintf("recipe-%d", id))
}

func recipeIdFromKeyName(key []byte) (blizzardv2.RecipeId, error) {
	unparsedId, err := strconv.Atoi(string(key)[len("recipe-"):])
	if err != nil {
		return blizzardv2.RecipeId(0), err
	}

	return blizzardv2.RecipeId(unparsedId), nil
}

func recipeNameKeyName(id blizzardv2.RecipeId) []byte {
	return []byte(fmt.Sprintf("recipe-name-%d", id))
}

func recipeIdFromNameKeyName(key []byte) (blizzardv2.RecipeId, error) {
	unparsedId, err := strconv.Atoi(string(key)[len("recipe-name-"):])
	if err != nil {
		return blizzardv2.RecipeId(0), err
	}

	return blizzardv2.RecipeId(unparsedId), nil
}

func itemsCraftedByKeyName(id blizzardv2.ItemId) []byte {
	return []byte(fmt.Sprintf("item-%d-crafted-by", id))
}

func isCompleteKeyName(flag professionsflags.ProfessionFlag) []byte {
	return []byte(flag)
}

// db
func DatabasePath(dbDir string) (string, error) {
	return fmt.Sprintf("%s/professions.db", dbDir), nil
}
