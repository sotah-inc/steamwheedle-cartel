package dev

func (sta DownloadAuctionsState) Run() error {
	if _, err := sta.Collector.CollectAuctions(); err != nil {
		return err
	}

	return nil
}
