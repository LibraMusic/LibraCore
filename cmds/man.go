package cmds

import (
	"fmt"
	"os"

	mcoral "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/spf13/cobra"
)

var manCmd = &cobra.Command{
	Use:                   "man",
	Short:                 "Generates Libra's command line manpages",
	SilenceUsage:          true,
	DisableFlagsInUseLine: true,
	Hidden:                true,
	Args:                  cobra.NoArgs,
	ValidArgsFunction:     cobra.NoFileCompletions,
	RunE: func(_ *cobra.Command, _ []string) error {
		manPage, err := mcoral.NewManPage(1, rootCmd.Root())
		if err != nil {
			return err
		}

		_, err = fmt.Fprint(os.Stdout, manPage.Build(roff.NewDocument()))
		return err
	},
}

func init() {
	rootCmd.AddCommand(manCmd)
}
