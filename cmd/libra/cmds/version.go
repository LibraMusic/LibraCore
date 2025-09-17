package cmds

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/libramusic/libracore"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version and build information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(libracore.GetVersionInfo())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
