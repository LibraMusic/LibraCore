package cmds

import (
	"charm.land/log/v2"
	"github.com/spf13/cobra"

	"github.com/libramusic/taurus/v2"

	"github.com/libramusic/libracore"
	"github.com/libramusic/libracore/config"
)

var (
	dataDir string
	l       *libracore.Libra
)

var rootCmd = &cobra.Command{
	Use:   "libra",
	Short: "Libra is a new, open, and extensible music service. Libra does what you want, how you want.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().
		StringVar(&dataDir, "dataDir", "", "persistent data directory (usually for containers)")
	_ = rootCmd.MarkFlagDirname("dataDir")

	rootCmd.PersistentFlags().String("logLevel", "", "log level (debug|info|warn|error)")
	_ = rootCmd.RegisterFlagCompletionFunc(
		"logLevel",
		cobra.FixedCompletions([]string{"debug", "info", "warn", "error"}, cobra.ShellCompDirectiveNoFileComp),
	)
	taurus.BindFlag("Logs.Level", rootCmd.PersistentFlags().Lookup("logLevel"))

	conf, err := config.LoadConfig(dataDir)
	if err != nil {
		log.Fatal("Failed to load config", "err", err)
	}
	libracore.SetupLogger(conf.Logs.Format, conf.Logs.Level)

	l, err = libracore.New(&conf)
	if err != nil {
		log.Fatal("Failed to create Libra instance", "err", err)
	}
}
