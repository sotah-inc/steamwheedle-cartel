package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (sta State) ResolveItemClasses(regionHostname string) (sotah.ItemClasses, error) {
	icComposite, err := blizzard.GetAllItemClasses(blizzard.GetAllItemClassesOptions{
		RegionHostname: regionHostname,
		GetItemClassIndexURL: func(providedRegionHostName string) (string, error) {
			uri, _ := sta.IO.Resolver.GetItemClassIndexURL(providedRegionHostName)

			appended, err := sta.IO.Resolver.AppendAccessToken(uri)
			if err != nil {
				return "", err
			}

			return appended, nil
		},
		GetItemClassURL: func(providedRegionHostName string, id int) (string, error) {
			uri, _ := sta.IO.Resolver.GetItemClassURL(providedRegionHostName, id)

			appended, err := sta.IO.Resolver.AppendAccessToken(uri)
			if err != nil {
				return "", err
			}

			return appended, nil
		},
	})
	if err != nil {
		return sotah.ItemClasses{}, err
	}

	out := sotah.ItemClasses{
		Classes: make([]sotah.ItemClass, len(icComposite)),
	}
	for i := 0; i < len(icComposite); i++ {
		sClasses := make([]sotah.SubItemClass, len(icComposite[i].ItemSubClasses))
		for j := 0; j < len(icComposite[i].ItemSubClasses); j += 1 {
			sClasses[j] = sotah.SubItemClass{
				SubClass: sotah.ItemSubClassClass(icComposite[i].ItemSubClasses[j].Id),
				Name:     icComposite[i].ItemSubClasses[j].Name,
			}
		}

		out.Classes[i] = sotah.ItemClass{
			Class:      sotah.ItemClassClass(icComposite[i].Id),
			Name:       icComposite[i].Name,
			SubClasses: nil,
		}
	}

	return out, nil
}
