package bus

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func NewItemIconBatchesMessages(batches sotah.IconItemsPayloadsBatches) ([]Message, error) {
	messages := []Message{}
	for i, ids := range batches {
		data, err := ids.EncodeForDelivery()
		if err != nil {
			return []Message{}, err
		}

		msg := NewMessage()
		msg.Data = data
		msg.ReplyToId = fmt.Sprintf("batch-%d", i)
		messages = append(messages, msg)
	}

	return messages, nil
}

func NewItemBatchesMessages(batches sotah.ItemIdBatches) ([]Message, error) {
	messages := []Message{}
	for i, ids := range batches {
		data, err := ids.EncodeForDelivery()
		if err != nil {
			return []Message{}, err
		}

		msg := NewMessage()
		msg.Data = data
		msg.ReplyToId = fmt.Sprintf("batch-%d", i)
		messages = append(messages, msg)
	}

	return messages, nil
}

func NewCollectAuctionMessages(regionRealms sotah.RegionRealms) ([]Message, error) {
	messages := []Message{}
	for _, realms := range regionRealms {
		for _, realm := range realms {
			job := CollectAuctionsJob{
				RegionName: string(realm.Region.Name),
				RealmSlug:  string(realm.Slug),
			}
			jsonEncoded, err := json.Marshal(job)
			if err != nil {
				return []Message{}, err
			}

			msg := NewMessage()
			msg.Data = string(jsonEncoded)
			msg.ReplyToId = fmt.Sprintf("%s-%s", realm.Region.Name, realm.Slug)
			messages = append(messages, msg)
		}
	}

	return messages, nil
}

func NewItemIdMessages(itemIds blizzard.ItemIds) []Message {
	messages := []Message{}
	for _, id := range itemIds {
		msg := NewMessage()
		msg.Data = strconv.Itoa(int(id))
		msg.ReplyToId = fmt.Sprintf("item-%d", id)
		messages = append(messages, msg)
	}

	return messages
}

func NewCleanupAuctionManifestJobsMessages(jobs CleanupAuctionManifestJobs) ([]Message, error) {
	messages := []Message{}
	for i, job := range jobs {
		data, err := job.EncodeForDelivery()
		if err != nil {
			return []Message{}, err
		}

		msg := NewMessage()
		msg.Data = data
		msg.ReplyToId = fmt.Sprintf("job-%d", i)
		messages = append(messages, msg)
	}

	return messages, nil
}

func NewCleanupPricelistPayloadsMessages(payload sotah.CleanupPricelistPayloads) ([]Message, error) {
	messages := []Message{}
	for i, payload := range payload {
		data, err := payload.EncodeForDelivery()
		if err != nil {
			return []Message{}, err
		}

		msg := NewMessage()
		msg.Data = data
		msg.ReplyToId = fmt.Sprintf("payload-%d", i)
		messages = append(messages, msg)
	}

	return messages, nil
}

func NewMessage() Message {
	return Message{Code: codes.Ok}
}

type Message struct {
	Data      string     `json:"data"`
	Err       string     `json:"error"`
	Code      codes.Code `json:"code"`
	ReplyTo   string     `json:"reply_to"`
	ReplyToId string     `json:"reply_to_id"`
}
