// Copyright © 2018 One Concern

package cmd

import (
	"github.com/spf13/cobra"
)

// bundleCmd represents the bundle command
var bundleCmd = &cobra.Command{
	Use:   "bundle",
	Short: "Commands to manage bundles for a repo",
	Long: `Commands to manage bundles for a repo.

A bundle is a group of files that were changed together.
Every bundle is an entry in the history of a repository.
`,
}

var bundleOptions struct {
	Id string
	CachePath string
	DownloadPath string
}

func init() {
	rootCmd.AddCommand(bundleCmd)
}

func addBundleFlag(cmd *cobra.Command) error {
	cmd.Flags().StringVarP(&bundleOptions.Id, "id", "i","", "The id for the bundle")
	return cmd.MarkFlagRequired("id")
}

func addCachePathFlag(cmd *cobra.Command) error {
	cmd.Flags().StringVarP(&bundleOptions.CachePath, "cache", "c", "", "The path to the cache folder")
}
