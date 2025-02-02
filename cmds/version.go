package cmds

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/libramusic/libracore/utils"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version and build information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(utils.GetVersionInfo()) //nolint:forbidigo // CLI response
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
