package cmd

import (
	"github.com/oneconcern/datamon/pkg/engine"
	"github.com/spf13/cobra"
)

// downloadBundleCmd is the command to download a specific bundle from Datamon and store it locally. The primary purpose
// is to get a readonly view for the data that is part of a bundle.
var downloadBundleCmd = &cobra.Command{
	Use:	 "download",
	Short: "Download a readonly, non-interactive view of the entire data that is part of a bundle",
	Long:  "Downloads all the data part of bundle for read purposes.",
	Run: func(cmd* cobra.Command, args []string) {
		_ := engine.DownloadBundle()
	},
}

func init() {
	addRepoFlag(downloadBundleCmd)
	addBundleFlag(downloadBundleCmd)
	bundleCmd.AddCommand(downloadBundleCmd)
}

