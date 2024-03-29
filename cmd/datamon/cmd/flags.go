// Copyright © 2018 One Concern

package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	context2 "github.com/oneconcern/datamon/pkg/context"

	"github.com/oneconcern/datamon/pkg/core"

	"github.com/oneconcern/datamon/pkg/model"
	"github.com/oneconcern/datamon/pkg/storage"
	"github.com/oneconcern/datamon/pkg/storage/gcs"
	"github.com/oneconcern/datamon/pkg/storage/localfs"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type flagsT struct {
	bundle struct {
		ID                string
		DataPath          string
		Message           string
		MountPath         string
		File              string
		Daemonize         bool
		Stream            bool
		FileList          string
		SkipOnError       bool
		ConcurrencyFactor int
		NameFilter        string
	}
	web struct {
		port int
	}
	label struct {
		Prefix string
		Name   string
	}
	context struct {
		Descriptor model.Context
	}
	repo struct {
		RepoName    string
		Description string
	}
	root struct {
		credFile string
		logLevel string
		cpuProf  bool
	}
	core struct {
		Config            string
		ConcurrencyFactor int
		BatchSize         int
	}
}

var datamonFlags = flagsT{}

func addBundleFlag(cmd *cobra.Command) string {
	bundleID := "bundle"
	if cmd != nil {
		cmd.Flags().StringVar(&datamonFlags.bundle.ID, bundleID, "", "The hash id for the bundle, if not specified the latest bundle will be used")
	}
	return bundleID
}

func addDataPathFlag(cmd *cobra.Command) string {
	destination := "destination"
	cmd.Flags().StringVar(&datamonFlags.bundle.DataPath, destination, "", "The path to the download dir")
	return destination
}

func addNameFilterFlag(cmd *cobra.Command) string {
	nameFilter := "name-filter"
	cmd.Flags().StringVar(&datamonFlags.bundle.NameFilter, nameFilter, "",
		"A regular expression (RE2) to match names of bundle entries.")
	return nameFilter
}

func addMountPathFlag(cmd *cobra.Command) string {
	mount := "mount"
	cmd.Flags().StringVar(&datamonFlags.bundle.MountPath, mount, "", "The path to the mount dir")
	return mount
}

func addPathFlag(cmd *cobra.Command) string {
	path := "path"
	cmd.Flags().StringVar(&datamonFlags.bundle.DataPath, path, "", "The path to the folder or bucket (gs://<bucket>) for the data")
	return path
}

func addCommitMessageFlag(cmd *cobra.Command) string {
	message := "message"
	cmd.Flags().StringVar(&datamonFlags.bundle.Message, message, "", "The message describing the new bundle")
	return message
}

func addFileListFlag(cmd *cobra.Command) string {
	fileList := "files"
	cmd.Flags().StringVar(&datamonFlags.bundle.FileList, fileList, "", "Text file containing list of files separated by newline.")
	return fileList
}

func addBundleFileFlag(cmd *cobra.Command) string {
	file := "file"
	cmd.Flags().StringVar(&datamonFlags.bundle.File, file, "", "The file to download from the bundle")
	return file
}

func addDaemonizeFlag(cmd *cobra.Command) string {
	daemonize := "daemonize"
	if cmd != nil {
		cmd.Flags().BoolVar(&datamonFlags.bundle.Daemonize, daemonize, false, "Whether to run the command as a daemonized process")
	}
	return daemonize
}

func addStreamFlag(cmd *cobra.Command) string {
	stream := "stream"
	cmd.Flags().BoolVar(&datamonFlags.bundle.Stream, stream, true, "Stream in the FS view of the bundle, do not download all files. Default to true.")
	return stream
}

func addSkipMissingFlag(cmd *cobra.Command) string {
	skipOnError := "skip-on-error"
	cmd.Flags().BoolVar(&datamonFlags.bundle.SkipOnError, skipOnError, false, "Skip files encounter errors while reading."+
		"The list of files is either generated or passed in. During upload files can be deleted or encounter an error. Setting this flag will skip those files. Default to false")
	return skipOnError
}

const concurrencyFactorFlag = "concurrency-factor"

func addConcurrencyFactorFlag(cmd *cobra.Command, defaultConcurrency int) string {
	concurrencyFactor := concurrencyFactorFlag
	cmd.Flags().IntVar(&datamonFlags.bundle.ConcurrencyFactor, concurrencyFactor, defaultConcurrency,
		"Heuristic on the amount of concurrency used by various operations.  "+
			"Turn this value down to use less memory, increase for faster operations.")
	return concurrencyFactor
}

func addCoreConcurrencyFactorFlag(cmd *cobra.Command, defaultConcurrency int) string {
	// this takes the usual "concurrency-factor" flag, but sets non-object specific settings
	concurrencyFactor := concurrencyFactorFlag
	cmd.Flags().IntVar(&datamonFlags.core.ConcurrencyFactor, concurrencyFactor, defaultConcurrency,
		"Heuristic on the amount of concurrency used by core operations. "+
			"Concurrent retrieval of metadata is capped by the 'batch-size' parameter. "+
			"Turn this value down to use less memory, increase for faster operations.")
	return concurrencyFactor
}
func addContextFlag(cmd *cobra.Command) string {
	c := "context"
	cmd.Flags().StringVar(&datamonFlags.context.Descriptor.Name, c, "dev", "Set the context for datamon")
	return c
}

func addConfigFlag(cmd *cobra.Command) string {
	config := "config"
	cmd.Flags().StringVar(&datamonFlags.core.Config, config, "", "Set the config to use")
	return config
}

func addBatchSizeFlag(cmd *cobra.Command) string {
	batchSize := "batch-size"
	cmd.Flags().IntVar(&datamonFlags.core.BatchSize, batchSize, 1024,
		"Number of bundles streamed together as a batch. This can be tuned for performance based on network connectivity")
	return batchSize
}

func addWebPortFlag(cmd *cobra.Command) string {
	cmd.Flags().IntVar(&datamonFlags.web.port, webPort, 3003, "Port number for the web server")
	return webPort
}

func addLabelNameFlag(cmd *cobra.Command) string {
	labelName := "label"
	cmd.Flags().StringVar(&datamonFlags.label.Name, labelName, "", "The human-readable name of a label")
	return labelName
}

func addLabelPrefixFlag(cmd *cobra.Command) string {
	prefixString := "prefix"
	cmd.Flags().StringVar(&datamonFlags.label.Prefix, prefixString, "", "List labels starting with a prefix.")
	return prefixString
}

func addRepoNameOptionFlag(cmd *cobra.Command) string {
	repo := "repo"
	cmd.Flags().StringVar(&datamonFlags.repo.RepoName, repo, "", "The name of this repository")
	return repo
}

func addRepoDescription(cmd *cobra.Command) string {
	description := "description"
	cmd.Flags().StringVar(&datamonFlags.repo.Description, description, "", "The description for the repo")
	return description
}

func addBlobBucket(cmd *cobra.Command) string {
	blob := "blob"
	cmd.Flags().StringVar(&datamonFlags.context.Descriptor.Blob, blob, "", "The name of the bucket hosting the datamon blobs")
	return blob
}

func addMetadataBucket(cmd *cobra.Command) string {
	meta := "meta"
	cmd.Flags().StringVar(&datamonFlags.context.Descriptor.Metadata, meta, "", "The name of the bucket used by datamon metadata")
	return meta
}

func addVMetadataBucket(cmd *cobra.Command) string {
	vm := "vmeta"
	cmd.Flags().StringVar(&datamonFlags.context.Descriptor.VMetadata, vm, "", "The name of the bucket hosting the versioned metadata")
	return vm
}

func addWALBucket(cmd *cobra.Command) string {
	b := "wal"
	cmd.Flags().StringVar(&datamonFlags.context.Descriptor.WAL, b, "", "The name of the bucket hosting the WAL")
	return b
}

func addReadLogBucket(cmd *cobra.Command) string {
	b := "read-log"
	cmd.Flags().StringVar(&datamonFlags.context.Descriptor.ReadLog, b, "", "The name of the bucket hosting the read log")
	return b
}

func addCredentialFile(cmd *cobra.Command) string {
	credential := "credential"
	cmd.Flags().StringVar(&datamonFlags.root.credFile, credential, "", "The path to the credential file")
	return credential
}

func addLogLevel(cmd *cobra.Command) string {
	loglevel := "loglevel"
	cmd.Flags().StringVar(&datamonFlags.root.logLevel, loglevel, "info", "The logging level")
	return loglevel
}

func addCPUProfFlag(cmd *cobra.Command) string {
	c := "cpuprof"
	cmd.Flags().BoolVar(&datamonFlags.root.cpuProf, c, false, "Toggle runtime profiling")
	return c
}

/** parameters struct to other formats */

func paramsToDatamonContext(ctx context.Context, params flagsT) (context2.Stores, error) {
	stores := context2.Stores{}

	meta, err := gcs.New(ctx, params.context.Descriptor.Metadata, config.Credential)
	if err != nil {
		return context2.Stores{}, fmt.Errorf("failed to initialize metadata store, err:%s", err)
	}
	stores.SetMetadata(meta)

	blob, err := gcs.New(ctx, params.context.Descriptor.Blob, config.Credential)
	if err != nil {
		return context2.Stores{}, fmt.Errorf("failed to initialize blob store, err:%s", err)
	}
	stores.SetBlob(blob)

	v, err := gcs.New(ctx, params.context.Descriptor.VMetadata, config.Credential)
	if err != nil {
		return context2.Stores{}, fmt.Errorf("failed to initialize vmetadata store, err:%s", err)
	}
	stores.SetVMetadata(v)

	w, err := gcs.New(ctx, params.context.Descriptor.WAL, config.Credential)
	if err != nil {
		return context2.Stores{}, fmt.Errorf("failed to initialize wal store, err:%s", err)
	}
	stores.SetWal(w)

	r, err := gcs.New(ctx, params.context.Descriptor.ReadLog, config.Credential)
	if err != nil {
		return context2.Stores{}, fmt.Errorf("failed to initialize read log store, err:%s", err)
	}
	stores.SetReadLog(r)

	return stores, nil
}

func paramsToBundleOpts(stores context2.Stores) []core.BundleOption {
	ops := []core.BundleOption{
		core.ContextStores(stores),
	}
	return ops
}

func paramsToSrcStore(ctx context.Context, params flagsT, create bool) (storage.Store, error) {
	var err error
	var consumableStorePath string

	switch {
	case params.bundle.DataPath == "":
		consumableStorePath, err = ioutil.TempDir("", "datamon-mount-destination")
		if err != nil {
			return nil, fmt.Errorf("couldn't create temporary directory: %w", err)
		}
	case strings.HasPrefix(params.bundle.DataPath, "gs://"):
		consumableStorePath = params.bundle.DataPath
	default:
		consumableStorePath, err = sanitizePath(params.bundle.DataPath)
		if err != nil {
			return nil, fmt.Errorf("failed to sanitize destination: %v: %w",
				params.bundle.DataPath, err)
		}
	}

	if create {
		createPath(consumableStorePath)
	}

	var sourceStore storage.Store
	if strings.HasPrefix(consumableStorePath, "gs://") {
		fmt.Println(consumableStorePath[4:])
		sourceStore, err = gcs.New(ctx, consumableStorePath[5:], config.Credential)
		if err != nil {
			return sourceStore, err
		}
	} else {
		DieIfNotAccessible(consumableStorePath)
		DieIfNotDirectory(consumableStorePath)
		sourceStore = localfs.New(afero.NewBasePathFs(afero.NewOsFs(), consumableStorePath))
	}
	return sourceStore, nil
}

// DestT defines the nomenclature for allowed destination types (e.g. Empty/NonEmpty)
type DestT uint

const (
	destTEmpty = iota
	destTMaybeNonEmpty
	destTNonEmpty
)

func paramsToDestStore(params flagsT,
	destT DestT,
	tmpdirPrefix string,
) (storage.Store, error) {
	var err error
	var consumableStorePath string
	var destStore storage.Store

	if tmpdirPrefix != "" && params.bundle.DataPath != "" {
		tmpdirPrefix = ""
	}

	if tmpdirPrefix != "" {
		if destT == destTNonEmpty {
			return nil, fmt.Errorf("can't specify temp dest path and non-empty dir mutually exclusive")
		}
		consumableStorePath, err = ioutil.TempDir("", tmpdirPrefix)
		if err != nil {
			return nil, fmt.Errorf("couldn't create temporary directory: %w", err)
		}
	} else {
		consumableStorePath, err = sanitizePath(params.bundle.DataPath)
		if err != nil {
			return nil, fmt.Errorf("failed to sanitize destination: %s: %w", params.bundle.DataPath, err)
		}
		createPath(consumableStorePath)
	}

	fs := afero.NewBasePathFs(afero.NewOsFs(), consumableStorePath)

	if destT == destTEmpty {
		var empty bool
		empty, err = afero.IsEmpty(fs, "/")
		if err != nil {
			return nil, fmt.Errorf("failed path validation: %w", err)
		}
		if !empty {
			return nil, fmt.Errorf("%s should be empty", consumableStorePath)
		}
	} else if destT == destTNonEmpty {
		/* fail-fast impl.  partial dupe of model pkg, encoded here to more fully encode intent
		 * of this package independently.
		 */
		var ok bool
		ok, err = afero.DirExists(fs, ".datamon")
		if err != nil {
			return nil, fmt.Errorf("failed to look for metadata dir: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("failed to find metadata dir in %v", consumableStorePath)
		}
	}
	destStore = localfs.New(fs)
	return destStore, nil
}

func paramsToContributor(_ flagsT) (model.Contributor, error) {
	return authorizer.Principal(config.Credential)
}
