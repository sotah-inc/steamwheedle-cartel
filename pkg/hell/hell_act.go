package hell

type ActEndpoints struct {
	DownloadAuctions string `firestore:"download_auctions"`
}

func (c Client) GetActEndpoints() (ActEndpoints, error) {
	endpointsRef, err := c.FirmDocument("connection_info/act_endpoints")
	if err != nil {
		return ActEndpoints{}, err
	}

	docsnap, err := endpointsRef.Get(c.Context)
	if err != nil {
		return ActEndpoints{}, err
	}

	var actEndpoints ActEndpoints
	if err := docsnap.DataTo(&actEndpoints); err != nil {
		return ActEndpoints{}, err
	}

	return actEndpoints, nil
}
