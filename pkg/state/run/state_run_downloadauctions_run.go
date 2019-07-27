package run

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah/codes"
)

func (sta DownloadAuctionsState) ResolveRealm(tuple sotah.RegionRealmTuple) (sotah.Realm, error) {
	region, err := func() (sotah.Region, error) {
		for _, reg := range sta.regions {
			if reg.Name == blizzard.RegionName(tuple.RegionName) {
				return reg, nil
			}
		}

		return sotah.Region{}, errors.New("could not resolve region from job")
	}()
	if err != nil {
		return sotah.Realm{}, err
	}

	realm := sotah.NewSkeletonRealm(blizzard.RegionName(tuple.RegionName), blizzard.RealmSlug(tuple.RealmSlug))
	realm.Region = region

	return realm, nil
}

func (sta DownloadAuctionsState) Handle(regionRealmTuple sotah.RegionRealmTuple) sotah.Message {
	realm, err := sta.ResolveRealm(regionRealmTuple)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	uri, err := sta.blizzardClient.AppendAccessToken(blizzard.DefaultGetAuctionInfoURL(
		realm.Region.Hostname,
		realm.Slug,
	))
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	aucInfo, respMeta, err := blizzard.NewAuctionInfoFromHTTP(uri)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}
	if respMeta.Status != http.StatusOK {
		respError := blizzard.ResponseError{
			Status: respMeta.Status,
			Body:   string(respMeta.Body),
			URI:    uri,
		}
		data, err := json.Marshal(respError)
		if err != nil {
			return sotah.NewErrorMessage(err)
		}

		out := sotah.NewErrorMessage(errors.New("response status for auc-info was not OK"))
		out.Code = codes.BlizzardError
		out.Data = string(data)

		return out
	}

	aucInfoFile, err := func() (blizzard.AuctionFile, error) {
		if len(aucInfo.Files) == 0 {
			return blizzard.AuctionFile{}, errors.New("auc-info files was blank")
		}

		return aucInfo.Files[0], nil
	}()
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	lastModifiedTime := aucInfoFile.LastModifiedAsTime()
	lastModifiedTimestamp := sotah.UnixTimestamp(lastModifiedTime.Unix())

	obj := sta.auctionsStoreBase.GetObject(realm, lastModifiedTime, sta.auctionsBucket)
	exists, err := sta.auctionsStoreBase.ObjectExists(obj)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}
	if exists {
		logging.WithFields(logrus.Fields{
			"region":        realm.Region.Name,
			"realm":         realm.Slug,
			"last-modified": lastModifiedTimestamp,
		}).Info("Object exists for region/ realm/ last-modified tuple, skipping")

		out := sotah.NewMessage()
		out.Code = codes.NoAction

		return out
	}

	resp, err := blizzard.Download(aucInfoFile.URL)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}
	if resp.Status != http.StatusOK {
		respError := blizzard.ResponseError{
			Status: resp.Status,
			Body:   string(resp.Body),
			URI:    aucInfoFile.URL,
		}
		data, err := json.Marshal(respError)
		if err != nil {
			return sotah.NewErrorMessage(err)
		}

		out := sotah.NewErrorMessage(errors.New("response status for aucs was not OK"))
		out.Code = codes.BlizzardError
		out.Data = string(data)

		return out
	}

	logging.WithFields(logrus.Fields{
		"region":              realm.Region.Name,
		"realm":               realm.Slug,
		"last-modified":       lastModifiedTimestamp,
		"ingested-size-bytes": resp.ContentLength,
	}).Info("Parsed, saving to raw-auctions store")
	if err := sta.auctionsStoreBase.Handle(resp.Body, lastModifiedTime, realm, sta.auctionsBucket); err != nil {
		return sotah.NewErrorMessage(err)
	}

	logging.WithFields(logrus.Fields{
		"region":        realm.Region.Name,
		"realm":         realm.Slug,
		"last-modified": lastModifiedTimestamp,
	}).Info("Saved, adding to auction-manifest file")
	if err := sta.auctionManifestStoreBase.Handle(lastModifiedTimestamp, realm, sta.auctionsManifestBucket); err != nil {
		return sotah.NewErrorMessage(err)
	}

	replyTuple := sotah.RegionRealmTimestampSizeTuple{
		RegionRealmTimestampTuple: sotah.RegionRealmTimestampTuple{
			RegionRealmTuple: regionRealmTuple,
			TargetTimestamp:  int(lastModifiedTimestamp),
		},
		SizeBytes: resp.ContentLength,
	}
	encodedReplyTuple, err := replyTuple.EncodeForDelivery()
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	out := sotah.NewMessage()
	out.Code = codes.Ok
	out.Data = encodedReplyTuple

	return out
}

func (sta DownloadAuctionsState) Run(data []byte) sotah.Message {
	regionRealmTuple, err := sotah.NewRegionRealmTuple(string(data))
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	return sta.Handle(regionRealmTuple)
}
