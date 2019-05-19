package web

import (
	"context"
	"io"
	"os"

	"cloud.google.com/go/storage"
)

// CreateBucket creates a new bucket on Google Storage.
func CreateBucket(bucket, project string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	return client.Bucket(bucket).Create(ctx, project, nil)
}

// StorageUpload uploads a local file to Google Storage bucket.
func StorageUpload(bucket, object, local string) (int64, error) {
	var size int64
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return size, err
	}
	f, err := os.Open(local)
	if err != nil {
		return size, err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return size, err
	}
	size = stat.Size()
	wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return size, err
	}
	if err := wc.Close(); err != nil {
		return size, err
	}
	return size, nil
}
