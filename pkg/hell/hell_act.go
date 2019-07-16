package hell

import (
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func (c Client) GetActEndpoints() (sotah.ActEndpoints, error) {
	endpointsRef, err := c.FirmDocument("")
	if err != nil {
		return sotah.ActEndpoints{}, err
	}

	docsnap, err := endpointsRef.Get(c.Context)
	if err != nil {
		return sotah.ActEndpoints{}, err
	}

	var actEndpoints sotah.ActEndpoints
	if err := docsnap.DataTo(&actEndpoints); err != nil {
		return sotah.ActEndpoints{}, err
	}

	return actEndpoints, nil
}
