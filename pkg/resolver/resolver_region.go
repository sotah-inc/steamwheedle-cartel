package resolver

import (
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func (r Resolver) NewStatus(reg sotah.Region) (sotah.Status, error) {
	url := r.GetStatusURL(reg.Hostname)
	appendedUrl, err := r.AppendAccessToken(url)
	if err != nil {
		return sotah.Status{}, err
	}

	logging.WithFields(logrus.Fields{
		"url":                   url,
		"url-with-access-token": appendedUrl,
		"region":                reg.Name,
	}).Info("gathering status for region")

	resp, err := r.Download(url, true)
	if err != nil {
		return sotah.Status{}, err
	}
	if resp.Status != http.StatusOK {
		return sotah.Status{}, errors.New("status was not 200")
	}

	stat, err := blizzard.NewStatus(resp.Body)
	if err != nil {
		return sotah.Status{}, err
	}

	return sotah.Status{Status: stat, Region: reg, Realms: sotah.NewRealms(reg, stat.Realms)}, nil
}

type GetStatusesJob struct {
	Err    error
	Region sotah.Region
	Status sotah.Status
}

func (r Resolver) GetStatuses(regions sotah.RegionList) chan GetStatusesJob {
	// establishing channels
	out := make(chan GetStatusesJob)
	in := make(chan sotah.Region)

	// spinning up the workers for fetching items
	worker := func() {
		for region := range in {
			status, err := r.NewStatus(region)
			if err != nil {
				out <- GetStatusesJob{
					Err:    err,
					Region: sotah.Region{},
					Status: sotah.Status{},
				}

				continue
			}

			out <- GetStatusesJob{
				Err:    nil,
				Region: region,
				Status: status,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	// queueing up the items
	go func() {
		for _, region := range regions {
			in <- region
		}

		close(in)
	}()

	return out
}
