package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type BlizzardState struct {
	BlizzardClient blizzardv2.Client
}

func (sta BlizzardState) ResolveRegionConnectedRealms(
	regions sotah.RegionList,
) (map[blizzardv2.RegionName][]blizzardv2.ConnectedRealmResponse, error) {
	out := map[blizzardv2.RegionName][]blizzardv2.ConnectedRealmResponse{}
	for _, region := range regions {
		var err error
		out[region.Name], err = sta.resolveConnectedRealms(region)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

func (sta BlizzardState) resolveConnectedRealms(region sotah.Region) ([]blizzardv2.ConnectedRealmResponse, error) {
	return blizzardv2.GetAllConnectedRealms(blizzardv2.GetAllConnectedRealmsOptions{
		GetConnectedRealmIndexURL: func() (string, error) {
			return sta.BlizzardClient.AppendAccessToken(
				blizzardv2.DefaultConnectedRealmIndexURL(region.Hostname, region.Name),
			)
		},
		GetConnectedRealmURL: sta.BlizzardClient.AppendAccessToken,
	})
}

func (sta BlizzardState) ResolveItemClasses(regions sotah.RegionList) ([]blizzardv2.ItemClassResponse, error) {
	primaryRegion, err := regions.GetPrimaryRegion()
	if err != nil {
		return []blizzardv2.ItemClassResponse{}, err
	}

	return blizzardv2.GetAllItemClasses(blizzardv2.GetAllItemClassesOptions{
		GetItemClassIndexURL: func() (string, error) {
			return sta.BlizzardClient.AppendAccessToken(
				blizzardv2.DefaultGetItemClassIndexURL(primaryRegion.Hostname, primaryRegion.Name),
			)
		},
		GetItemClassURL: func(id blizzardv2.ItemClassId) (string, error) {
			return sta.BlizzardClient.AppendAccessToken(
				blizzardv2.DefaultGetItemClassURL(primaryRegion.Hostname, primaryRegion.Name, id),
			)
		},
	})
}

func (sta BlizzardState) ResolveTokens(
	regions sotah.RegionList,
) (map[blizzardv2.RegionName]blizzardv2.TokenResponse, error) {
	return blizzardv2.GetTokens(blizzardv2.GetTokensOptions{
		Tuples: func() []blizzardv2.RegionHostnameTuple {
			out := make([]blizzardv2.RegionHostnameTuple, len(regions))
			for i, region := range regions {
				out[i] = blizzardv2.RegionHostnameTuple{
					RegionName:     region.Name,
					RegionHostname: region.Hostname,
				}
			}

			return out
		}(),
		GetTokenInfoURL: func(regionHostname string, regionName blizzardv2.RegionName) (string, error) {
			return sta.BlizzardClient.AppendAccessToken(
				blizzardv2.DefaultGetTokenURL(regionHostname, regionName),
			)
		},
	})
}

func (sta BlizzardState) ResolveAuctions(tuples []blizzardv2.DownloadConnectedRealmTuple) chan blizzardv2.GetAuctionsJob {
	return blizzardv2.GetAuctions(blizzardv2.GetAuctionsOptions{
		Tuples: tuples,
		GetAuctionsURL: func(tuple blizzardv2.DownloadConnectedRealmTuple) (s string, err error) {
			return sta.BlizzardClient.AppendAccessToken(blizzardv2.DefaultGetAuctionsURL(tuple))
		},
	})
}
