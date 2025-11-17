package cmds

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/libramusic/libracore/db"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate the database",
	Long: `Migrate the database.
Uses your database connection string from the config file.`,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		err := db.Connect()
		if err != nil {
			return fmt.Errorf("database connection failed: %w", err)
		}
		log.Info("Connected to database", "engine", db.DB.EngineName())
		return nil
	},
	PersistentPostRunE: func(_ *cobra.Command, _ []string) error {
		if db.DB != nil {
			err := db.DB.Close()
			if err != nil {
				return fmt.Errorf("error closing database connection: %w", err)
			}
			log.Info("Database connection closed")
		}
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
		err := db.DB.MigrateUp(steps)
		if err != nil {
			log.Fatal("Error migrating database", "err", err)
		}
		fmt.Println("Database migration complete")
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
		err := db.DB.MigrateDown(steps)
		if err != nil {
			log.Fatal("Error migrating database", "err", err)
		}
		fmt.Println("Database migration complete")
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(upCmd)
	migrateCmd.AddCommand(downCmd)
}
