package blizzardv2

import (
	"regexp"
	"strconv"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetAllConnectedRealmsOptions struct {
	GetConnectedRealmIndexURL func() (string, error)
	GetConnectedRealmURL      func(string) (string, error)
	Blacklist                 []ConnectedRealmId
	RealmWhitelist            RealmSlugs
}

type GetAllConnectedRealmsJob struct {
	Err                    error
	HrefReference          HrefReference
	ConnectedRealmResponse ConnectedRealmResponse
}

func (job GetAllConnectedRealmsJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"href":  job.HrefReference.Href,
	}
}

func GetAllConnectedRealms(
	opts GetAllConnectedRealmsOptions,
) (chan GetAllConnectedRealmsJob, error) {
	// querying index
	uri, err := opts.GetConnectedRealmIndexURL()
	if err != nil {
		return nil, err
	}

	crIndex, _, err := NewConnectedRealmIndexFromHTTP(uri)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to get connected-realm-index")

		return nil, err
	}

	// starting up workers for gathering individual connected-realms
	in := make(chan HrefReference)
	out := make(chan GetAllConnectedRealmsJob)
	worker := func() {
		for hrefRef := range in {
			getConnectedRealmUri, err := opts.GetConnectedRealmURL(hrefRef.Href)
			if err != nil {
				out <- GetAllConnectedRealmsJob{
					Err:                    err,
					HrefReference:          hrefRef,
					ConnectedRealmResponse: ConnectedRealmResponse{},
				}

				continue
			}

			cRealm, _, err := NewConnectedRealmResponseFromHTTP(getConnectedRealmUri)
			if err != nil {
				out <- GetAllConnectedRealmsJob{
					Err:                    err,
					HrefReference:          hrefRef,
					ConnectedRealmResponse: ConnectedRealmResponse{},
				}

				continue
			}

			cRealm.Realms = func() RealmResponses {
				foundRealms := RealmResponses{}
				for _, realm := range cRealm.Realms {
					if len(opts.RealmWhitelist) > 0 && !opts.RealmWhitelist.Has(realm.Slug) {
						continue
					}

					foundRealms = append(foundRealms, realm)
				}

				return foundRealms
			}()

			if len(cRealm.Realms) == 0 {
				continue
			}

			out <- GetAllConnectedRealmsJob{
				Err:                    nil,
				HrefReference:          hrefRef,
				ConnectedRealmResponse: cRealm,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	// producing a regex for parsing connected-realm href
	re, err := regexp.Compile(`^.+/([0-9]+)\?.+$`)
	if err != nil {
		return nil, err
	}

	// producing a blacklist map
	blacklistMap := map[ConnectedRealmId]struct{}{}
	for _, id := range opts.Blacklist {
		blacklistMap[id] = struct{}{}
	}

	// queueing it up
	go func() {
		for _, hrefRef := range crIndex.ConnectedRealms {
			matches := re.FindStringSubmatch(hrefRef.Href)
			if len(matches) != 2 {
				logging.WithField(
					"matches",
					matches,
				).Error("regex match on href match count was not 2")

				continue
			}

			parsedId, err := strconv.Atoi(matches[1])
			if err != nil {
				logging.WithFields(logrus.Fields{
					"match-id": matches[1],
					"error":    err.Error(),
				}).Error("failed to parse match-id")

				continue
			}

			if _, ok := blacklistMap[ConnectedRealmId(parsedId)]; ok {
				logging.WithField("parsed-id", parsedId).Debug("skipping connected-realm")

				continue
			}

			in <- hrefRef
		}

		close(in)
	}()

	return out, nil
}
