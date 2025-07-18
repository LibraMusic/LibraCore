package cmds

import (
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/libramusic/taurus/v2"

	"github.com/libramusic/libracore/config"
	"github.com/libramusic/libracore/utils"
)

var rootCmd = &cobra.Command{
	Use:   "libra",
	Short: "Libra is a new, open, and extensible music service. Libra does what you want, how you want.",
	Long:  `Libra is a new, open, and extensible music service. Libra does what you want, how you want.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().
		StringVar(&config.DataDir, "dataDir", "", "persistent data directory (usually for containers)")
	_ = rootCmd.MarkFlagDirname("dataDir")

	rootCmd.PersistentFlags().String("logLevel", "", "log level (debug|info|warn|error)")
	_ = rootCmd.RegisterFlagCompletionFunc(
		"logLevel",
		cobra.FixedCompletions([]string{"debug", "info", "warn", "error"}, cobra.ShellCompDirectiveNoFileComp),
	)
	taurus.BindFlag("Logs.Level", rootCmd.PersistentFlags().Lookup("logLevel"))

	if err := config.LoadConfig(); err != nil {
		log.Fatal("Failed to load config", "err", err)
	}
	utils.SetupLogger(config.Conf.Logs.Format, config.Conf.Logs.Level)
}
