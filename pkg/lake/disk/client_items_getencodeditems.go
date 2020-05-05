package disk

import (
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type getEncodedItemJob struct {
	err                   error
	id                    blizzardv2.ItemId
	encodedItem           []byte
	encodedNormalizedName []byte
}

func (g getEncodedItemJob) Err() error                    { return g.err }
func (g getEncodedItemJob) Id() blizzardv2.ItemId         { return g.id }
func (g getEncodedItemJob) EncodedItem() []byte           { return g.encodedItem }
func (g getEncodedItemJob) EncodedNormalizedName() []byte { return g.encodedNormalizedName }
func (g getEncodedItemJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": g.err.Error(),
		"id":    g.id,
	}
}

func (client Client) GetEncodedItems(ids blizzardv2.ItemIds) chan BaseLake.GetEncodedItemJob {
	out := make(chan BaseLake.GetEncodedItemJob)

	// starting up workers for resolving items
	itemsOut := client.resolveItems(ids)

	// starting up workers for resolving item-medias
	itemMediasIn := make(chan blizzardv2.GetItemMediasInJob)
	itemMediasOut := client.resolveItemMedias(itemMediasIn)

	// queueing it all up
	go func() {
		for job := range itemsOut {
			if job.Err != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to resolve item")

				continue
			}

			logging.WithField("item-id", job.Id).Info("enqueueing item for item-media resolution")

			itemMediasIn <- blizzardv2.GetItemMediasInJob{Item: job.ItemResponse}
		}

		close(itemMediasIn)
	}()
	go func() {
		for job := range itemMediasOut {
			if job.Err != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to resolve item-media")

				continue
			}

			itemIcon, err := job.ItemMediaResponse.GetIcon()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":    err.Error(),
					"response": job.ItemMediaResponse,
				}).Error("failed to resolve item-icon from item-media")

				continue
			}

			logging.WithField("item-id", job.Item.Id).Info("enqueueing item for persistence")

			normalizedName, err := func() (locale.Mapping, error) {
				foundName := job.Item.Name
				if _, ok := foundName[locale.EnUS]; !ok {
					return locale.Mapping{}, errors.New("failed to resolve enUS name")
				}

				foundName[locale.EnUS], err = sotah.NormalizeString(foundName[locale.EnUS])
				if err != nil {
					return locale.Mapping{}, err
				}

				return foundName, nil
			}()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":    err.Error(),
					"response": job.ItemMediaResponse,
				}).Error("failed to normalize name")

				continue
			}

			item := sotah.Item{
				BlizzardMeta: job.Item,
				SotahMeta: sotah.ItemMeta{
					LastModified:   sotah.UnixTimestamp(time.Now().Unix()),
					NormalizedName: normalizedName,
					ItemIconMeta: sotah.ItemIconMeta{
						URL:        blizzardv2.DefaultGetItemIconURL(itemIcon),
						ObjectName: sotah.NewItemObjectName(sotah.IconName(itemIcon)),
						Icon:       sotah.IconName(itemIcon),
					},
				},
			}

			encodedItem, err := item.EncodeForStorage()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error": err.Error(),
					"item":  item.BlizzardMeta.Id,
				}).Error("failed to encode item for storage")

				continue
			}

			encodedNormalizedName, err := normalizedName.EncodeForStorage()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error": err.Error(),
					"item":  item.BlizzardMeta.Id,
				}).Error("failed to encode normalized-name for storage")

				continue
			}

			out <- getEncodedItemJob{
				id:                    job.Item.Id,
				encodedItem:           encodedItem,
				encodedNormalizedName: encodedNormalizedName,
			}
		}

		close(out)
	}()

	return out
}
