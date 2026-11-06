package cmds

import (
	"fmt"
	"strconv"

	"charm.land/log/v2"
	"github.com/spf13/cobra"
)

var migrateDatabase string

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate the database",
	Long: `Migrate the database.
Uses your database connection string from the config file.`,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		if err := l.Connect(); err != nil {
			return err
		}
		log.Info("Connected to databases")
		return nil
	},
	PersistentPostRunE: func(_ *cobra.Command, _ []string) error {
		if err := l.Close(); err != nil {
			return fmt.Errorf("error closing database connections: %w", err)
		}
		log.Info("Database connections closed")
		return nil
	},
}

var upCmd = &cobra.Command{
	Use:   "up [steps]",
	Short: "Migrate the database up",
	Long: `Migrate the database up.
Use 'steps' to specify the number of steps to migrate.`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: cobra.NoFileCompletions,
	Run: func(_ *cobra.Command, args []string) {
		steps := -1
		if len(args) > 0 {
			var err error
			steps, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Println("Error: steps must be an integer")
				return
			}
		}

		if migrateDatabase == "media" || migrateDatabase == "all" {
			err := l.MediaDB.MigrateUp(steps)
			if err != nil {
				log.Fatal("Error migrating media database", "err", err)
			}
		}

		if migrateDatabase == "auth" || (migrateDatabase == "all" && l.AuthDB != l.MediaDB) {
			err := l.AuthDB.MigrateUp(steps)
			if err != nil {
				log.Fatal("Error migrating auth database", "err", err)
			}
		}

		fmt.Println("Database migrations complete")
	},
}

var downCmd = &cobra.Command{
	Use:   "down [steps]",
	Short: "Migrate the database down",
	Long: `Migrate the database down.
Use 'steps' to specify the number of steps to migrate.`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: cobra.NoFileCompletions,
	Run: func(_ *cobra.Command, args []string) {
		steps := -1
		if len(args) > 0 {
			var err error
			steps, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Println("Error: steps must be an integer")
				return
			}
		}

		if migrateDatabase == "media" || migrateDatabase == "all" {
			err := l.MediaDB.MigrateDown(steps)
			if err != nil {
				log.Fatal("Error migrating media database", "err", err)
			}
		}

		if migrateDatabase == "auth" || (migrateDatabase == "all" && l.AuthDB != l.MediaDB) {
			err := l.AuthDB.MigrateDown(steps)
			if err != nil {
				log.Fatal("Error migrating auth database", "err", err)
			}
		}

		fmt.Println("Database migrations complete")
	},
}

func init() {
	migrateCmd.PersistentFlags().
		StringVarP(&migrateDatabase, "database", "d", "all", "database to migrate (media, auth, or all)")
	_ = migrateCmd.RegisterFlagCompletionFunc(
		"database",
		cobra.FixedCompletions([]string{"media", "auth", "all"}, cobra.ShellCompDirectiveNoFileComp),
	)

	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(upCmd)
	migrateCmd.AddCommand(downCmd)
}
