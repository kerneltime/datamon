package engine

import (
	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)
import "golang.org/x/net/context"

// Implements the functions for downloading and uploading Bundles
type BundleDownload struct {
	Repo 		string
	Bucket  string
	Bundle  string
	Branch  string
	CachePath string
	BundleFSPath string
}

func DownloadBundle() error {
	// Check for the presence of required folders
	// Download the bundle json
	// Read the file JSON
	// Download the blobs
	// Link the file
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithScopes(storage.ScopeReadOnly))
}
