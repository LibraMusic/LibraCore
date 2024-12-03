package cmds

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/LibraMusic/LibraCore/config"
	"github.com/LibraMusic/LibraCore/utils"
)

var rootCmd = &cobra.Command{
	Use:   "libra",
	Short: "Libra is a new, open, and extensible music service. Libra does what you want, how you want.",
	Long:  `Libra is a new, open, and extensible music service. Libra does what you want, how you want.`,
}

func Execute() {
	rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&utils.DataDir, "dataDir", "", "persistent data directory (usually for containers)")
	rootCmd.MarkFlagDirname("dataDir")
	if utils.DataDir == "" {
		utils.DataDir = os.Getenv("LIBRA_DATA_DIR")
	}

	rootCmd.PersistentFlags().String("logLevel", "", "log level (debug|info|warn|error)")
	rootCmd.RegisterFlagCompletionFunc("logLevel", cobra.FixedCompletions([]string{"debug", "info", "warn", "error"}, cobra.ShellCompDirectiveNoFileComp))
	viper.BindPFlag("logLevel", rootCmd.PersistentFlags().Lookup("logLevel"))
}

func initConfig() {
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config", "err", err)
	}
	config.Conf = conf
}
