package cmds

import (
	"fmt"

	"github.com/LibraMusic/LibraCore/utils"
	"github.com/spf13/cobra"
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
