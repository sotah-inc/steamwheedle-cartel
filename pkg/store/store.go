package store

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func NewClient(projectID string) (Client, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return Client{}, err
	}

	s := Client{Context: ctx, projectID: projectID, client: client}

	return s, nil
}

type Client struct {
	Context   context.Context
	projectID string
	client    *storage.Client
}

type DeleteAllFromTimestampsJob struct {
	sotah.RegionRealmTimestampTuple
	Err  error
	Size int64
}

type DeleteAllResults struct {
	TotalCount int
	TotalSize  int64
}

const itemIconURLFormat = "https://storage.googleapis.com/%s/%s"
