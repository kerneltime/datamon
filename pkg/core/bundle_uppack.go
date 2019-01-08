// Copyright © 2018 One Concern

package core

import (
	"context"

	"github.com/oneconcern/datamon/pkg/cafs"
	"github.com/oneconcern/datamon/pkg/model"
	"github.com/oneconcern/datamon/pkg/storage"
	"gopkg.in/yaml.v2"
)

func unpackBundleDescriptor(ctx context.Context, bundle *Bundle) error {

	bundleDescriptorBuffer, err := storage.ReadTee(ctx,
		bundle.ArchiveStore, model.GetArchivePathToBundle(bundle.RepoID, bundle.BundleID),
		bundle.ConsumableStore, model.GetConsumablePathToBundle(bundle.BundleID))
	if err != nil {
		return err
	}

	// Unmarshal the file
	err = yaml.Unmarshal(bundleDescriptorBuffer, &bundle.BundleDescriptor)
	if err != nil {
		return err
	}
	return nil
}

func unpackBundleFileList(ctx context.Context, bundle *Bundle) error {
	// Download the files json
	var i uint64
	for i = 0; i < bundle.BundleDescriptor.BundleEntriesFileCount; i++ {
		bundleEntriesBuffer, err := storage.ReadTee(ctx,
			bundle.ArchiveStore, model.GetArchivePathToBundleFileList(bundle.RepoID, bundle.BundleID, i),
			bundle.ConsumableStore, model.GetConsumablePathToBundleFileList(bundle.BundleID, uint64(i)))
		if err != nil {
			return err
		}
		var bundleEntries model.BundleEntries
		err = yaml.Unmarshal(bundleEntriesBuffer, &bundleEntries)
		if err != nil {
			return err
		}
		bundle.BundleEntries = append(bundle.BundleEntries, bundleEntries.BundleEntries...)
	}
	// Link the file
	return nil
}

func unpackDataFiles(ctx context.Context, bundle *Bundle) error {
	fs, err := cafs.New(
		cafs.LeafSize(bundle.BundleDescriptor.LeafSize),
		cafs.Backend(bundle.ArchiveStore),
		cafs.Prefix(model.GetArchivePathBlobPrefix()),
	)
	if err != nil {
		return err
	}
	for _, bundleEntry := range bundle.BundleEntries {
		key, err := cafs.KeyFromString(bundleEntry.Hash)
		if err != nil {
			return err
		}
		reader, err := fs.Get(ctx, key)
		if err != nil {
			return err
		}
		err = bundle.ConsumableStore.Put(ctx, bundleEntry.NameWithPath, reader)
		if err != nil {
			return err
		}
	}
	return nil
}
