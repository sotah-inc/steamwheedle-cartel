package hell

type ActEndpoints struct {
	DownloadAuctions    string `firestore:"download_auctions"`
	ComputeLiveAuctions string `firestore:"compute_live_auctions"`
	Gateway             string `firestore:"gateway"`
	CleanupManifests    string `firestore:"cleanup_manifests"`
	CleanupAuctions     string `firestore:"cleanup_auctions"`
	SyncItems           string `firestore:"sync_items"`
	SyncItemIcons       string `firestore:"sync_item_icons"`
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
