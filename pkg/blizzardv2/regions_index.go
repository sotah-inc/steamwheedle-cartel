package blizzardv2

import "fmt"

const regionIndexURLFormat = "https://%s/data/wow/region/index?namespace=dynamic-%s"

func DefaultRegionIndexURL(regionHostname string, regionName RegionName) string {
	return fmt.Sprintf(regionIndexURLFormat, regionHostname, regionName)
}

type GetRegionIndexURLFunc func(string) string

type RegionIndexResponse struct {
	LinksBase
}
