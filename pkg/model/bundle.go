package model

import (
	"fmt"
	"os"
	"time"
)

// Bundle represents a commit which is a file tree with the changes to the repository.
type Bundle struct {
	LeafSize        uint32        `json:"leafSize" yaml:"leafSize"` // Each bundles blobs are independently generated
	ID              string        `json:"id" yaml:"id"`
	Message         string        `json:"message" yaml:"message"`
	Parents         []string      `json:"parents,omitempty" yaml:"parents,omitempty"`
	Timestamp       time.Time     `json:"timestamp,omitempty" yaml:"timestamp,omitempty"`
	Committers      []Contributor `json:"committers" yaml:"committers"`
	EntryFilesCount int64         `json:"entryfilescount" yaml:"entryfilescount"`
	_               struct{}      `json:"-" yaml:"-"`
}

// List of all files part of a bundle.
type BundleEntries struct {
	BundleEntries []BundleEntry `json:"BundleEntries" yaml:"BundleEntries"`
	_             struct{}      `json:"-" yaml:"-"`
}

// List of files, directories (empty) skipped
type BundleEntry struct {
	Hash         string      `json:"hash" yaml:"hash"`
	NameWithPath string      `json:"name" yaml:"name"`
	FileMode     os.FileMode `json:"mode" yaml:"mode"`
	Size         uint        `json:"size" yaml:"size"`
	_            struct{}    `json:"-" yaml:"-"`
}

// Contributor who created the object
type Contributor struct {
	Name  string   `json:"name" yaml:"name"`
	Email string   `json:"email" yaml:"email"`
	_     struct{} `json:"-" yaml:"-"`
}

func (c *Contributor) String() string {
	if c.Email == "" {
		return c.Name
	}
	if c.Name == "" {
		return c.Email
	}
	return fmt.Sprintf("%s <%s>", c.Name, c.Email)
}

func GetConsumablePathToBundle(bundleId string) string {
	return fmt.Sprint("./.datamon/", bundleId, ".json")
}

func GetConsumablePathToBundleFileList(bundleId string, index int64) string {
	return fmt.Sprint("./.datamon/", bundleId, "-bundle-files-", index, ".json")
}

func GetArchivePathToBundle(repo string, bundleId string) string {
	return fmt.Sprint(repo, "-bundles/", bundleId, "/bundle.json")
}

func GetArchivePathToBundleFileList(repo string, bundleId string, index int64) string {
	// <repo>-bundles/<bundle>/bundlefiles-<index>.json
	return fmt.Sprint(repo, "-bundles/", bundleId, "/bundle-files-", index, ".json")
}

func GetArchivePathBlobPrefix() string {
	return "blobs/"
}

func GetBundleTimeStamp() time.Time {
	t := time.Now()
	return t.UTC()
}
