package hell

type ConnectionInfo struct {
	DbConnectionName    string `firestore:"db_connection_name"`
	DbHost              string `firestore:"db_host"`
	DbPassword          string `firestore:"db_password"`
	DownloadAuctionsURI string `firestore:"download-auctions-uri"`
	NatsHost            string `firestore:"nats_host"`
	NatsPort            string `firestore:"nats_port"`
}

func (c Client) GetConnectionInfo() (ConnectionInfo, error) {
	ref, err := c.FirmDocument("connection_info/current")
	if err != nil {
		return ConnectionInfo{}, err
	}

	docsnap, err := ref.Get(c.Context)
	if err != nil {
		return ConnectionInfo{}, err
	}

	var out ConnectionInfo
	if err := docsnap.DataTo(&out); err != nil {
		return ConnectionInfo{}, err
	}

	return out, nil
}
