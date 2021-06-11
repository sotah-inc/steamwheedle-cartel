package disk

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemclass"

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
	itemClass             itemclass.Id
	isVendorItem          bool
	vendorPrice           blizzardv2.PriceValue
}

func (g getEncodedItemJob) Err() error                         { return g.err }
func (g getEncodedItemJob) Id() blizzardv2.ItemId              { return g.id }
func (g getEncodedItemJob) EncodedItem() []byte                { return g.encodedItem }
func (g getEncodedItemJob) EncodedNormalizedName() []byte      { return g.encodedNormalizedName }
func (g getEncodedItemJob) ItemClass() itemclass.Id            { return g.itemClass }
func (g getEncodedItemJob) IsVendorItem() bool                 { return g.isVendorItem }
func (g getEncodedItemJob) VendorPrice() blizzardv2.PriceValue { return g.vendorPrice }
func (g getEncodedItemJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": g.err.Error(),
		"id":    g.id,
	}
}

type erroneousItemJob struct {
	statusCode int
	id         blizzardv2.ItemId
}

func (client Client) GetEncodedItems(
	version gameversion.GameVersion,
	ids blizzardv2.ItemIds,
) (chan BaseLake.GetEncodedItemJob, chan []blizzardv2.ItemId) {
	out := make(chan BaseLake.GetEncodedItemJob)

	// starting up workers for resolving items
	itemsOut := client.resolveItems(version, ids)

	// starting up workers for resolving item-medias
	itemMediasIn := make(chan blizzardv2.GetItemMediasInJob)
	itemMediasOut := client.resolveItemMedias(itemMediasIn)

	// starting up a worker for gathering erroneous status codes
	erroneousItemsIn := make(chan erroneousItemJob)
	erroneousItemIdsOut := make(chan []blizzardv2.ItemId)
	go func() {
		var erroneousItemIds []blizzardv2.ItemId
		for job := range erroneousItemsIn {
			if job.statusCode != http.StatusNotFound {
				continue
			}

			erroneousItemIds = append(erroneousItemIds, job.id)
		}

		erroneousItemIdsOut <- erroneousItemIds
		close(erroneousItemIdsOut)
	}()

	// queueing it all up
	go func() {
		for job := range itemsOut {
			if job.Err != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to resolve item")
				erroneousItemsIn <- erroneousItemJob{
					statusCode: job.Status,
					id:         job.Id,
				}

				continue
			}

			logging.WithField("item-id", job.Id).Info("enqueueing item for item-media resolution")

			itemMediasIn <- blizzardv2.GetItemMediasInJob{Item: job.ItemResponse}
		}

		close(itemMediasIn)
		close(erroneousItemsIn)
	}()
	go func() {
		for job := range itemMediasOut {
			if job.Err != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to resolve item-media")
				out <- getEncodedItemJob{
					err:                   job.Err,
					id:                    job.Item.Id,
					encodedItem:           []byte{},
					encodedNormalizedName: []byte{},
					itemClass:             0,
					isVendorItem:          false,
					vendorPrice:           0,
				}

				continue
			}

			itemIcon, err := job.ItemMediaResponse.GetIcon()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":    err.Error(),
					"response": job.ItemMediaResponse,
				}).Error("failed to resolve item-icon from item-media")
				out <- getEncodedItemJob{
					err:                   err,
					id:                    job.Item.Id,
					encodedItem:           []byte{},
					encodedNormalizedName: []byte{},
					itemClass:             0,
					isVendorItem:          false,
					vendorPrice:           0,
				}

				continue
			}

			normalizedName, err := func() (locale.Mapping, error) {
				foundName, ok := job.Item.Name[locale.EnUS]
				if !ok {
					return locale.Mapping{}, errors.New("failed to resolve enUS name")
				}

				normalizedName, err := sotah.NormalizeString(foundName)
				if err != nil {
					return locale.Mapping{}, err
				}

				return locale.Mapping{locale.EnUS: normalizedName}, nil
			}()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":    err.Error(),
					"response": job.ItemMediaResponse,
				}).Error("failed to normalize name")
				out <- getEncodedItemJob{
					err:                   err,
					id:                    job.Item.Id,
					encodedItem:           []byte{},
					encodedNormalizedName: []byte{},
					itemClass:             0,
					isVendorItem:          false,
					vendorPrice:           0,
				}

				continue
			}

			item := sotah.Item{
				BlizzardMeta: job.Item,
				SotahMeta: sotah.ItemMeta{
					LastModified:   sotah.UnixTimestamp(time.Now().Unix()),
					NormalizedName: normalizedName,
					ItemIconMeta: sotah.ItemIconMeta{
						URL:        blizzardv2.DefaultGetItemIconURL(itemIcon),
						ObjectName: "",
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
				out <- getEncodedItemJob{
					err:                   err,
					id:                    job.Item.Id,
					encodedItem:           []byte{},
					encodedNormalizedName: []byte{},
					itemClass:             0,
					isVendorItem:          false,
					vendorPrice:           0,
				}

				continue
			}

			encodedNormalizedName, err := normalizedName.EncodeForStorage()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error": err.Error(),
					"item":  item.BlizzardMeta.Id,
				}).Error("failed to encode normalized-name for storage")
				out <- getEncodedItemJob{
					err:                   err,
					id:                    job.Item.Id,
					encodedItem:           []byte{},
					encodedNormalizedName: []byte{},
					itemClass:             0,
					isVendorItem:          false,
					vendorPrice:           0,
				}

				continue
			}

			isVendorItem := strings.Contains(job.Item.Description.ResolveDefaultName(), "Sold by")

			if isVendorItem {
				logging.WithFields(logrus.Fields{
					"item":         job.Item.Id,
					"name":         job.Item.Name.ResolveDefaultName(),
					"description":  job.Item.Description.ResolveDefaultName(),
					"vendor-price": job.Item.PurchasePrice,
				}).Info("found item sold by vendor")
			}

			out <- getEncodedItemJob{
				err:                   nil,
				id:                    job.Item.Id,
				encodedItem:           encodedItem,
				encodedNormalizedName: encodedNormalizedName,
				itemClass:             job.Item.ItemClass.Id,
				isVendorItem:          isVendorItem,
				vendorPrice:           job.Item.PurchasePrice,
			}
		}

		close(out)
	}()

	return out, erroneousItemIdsOut
}
