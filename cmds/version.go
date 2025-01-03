package cmds

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/libramusic/libracore/utils"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version and build information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.GetVersionInfo())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
