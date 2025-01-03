package cmds

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/libramusic/taurus"
	"github.com/spf13/cobra"

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
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&utils.DataDir, "dataDir", "", "persistent data directory (usually for containers)")
	_ = rootCmd.MarkFlagDirname("dataDir")
	if utils.DataDir == "" {
		utils.DataDir = os.Getenv("LIBRA_DATA_DIR")
	}

	rootCmd.PersistentFlags().String("logLevel", "", "log level (debug|info|warn|error)")
	_ = rootCmd.RegisterFlagCompletionFunc("logLevel", cobra.FixedCompletions([]string{"debug", "info", "warn", "error"}, cobra.ShellCompDirectiveNoFileComp))
	taurus.BindFlag("Logs.LogLevel", rootCmd.PersistentFlags().Lookup("logLevel"))
}

func initConfig() {
	if err := config.LoadConfig(); err != nil {
		log.Fatal("Failed to load config", "err", err)
	}
}
