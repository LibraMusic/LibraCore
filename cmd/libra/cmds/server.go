package cmds

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/markbates/goth"
	"github.com/spf13/cobra"

	"github.com/libramusic/taurus/v2"

	"github.com/libramusic/libracore/api"
	"github.com/libramusic/libracore/api/metrics"
	"github.com/libramusic/libracore/api/routes/auth"
	"github.com/libramusic/libracore/api/server"
	"github.com/libramusic/libracore/config"
	"github.com/libramusic/libracore/db"
	"github.com/libramusic/libracore/storage"
)

var serverCmd = &cobra.Command{
	Use:               "server",
	Aliases:           []string{"start"},
	Short:             "Start the server",
	Args:              cobra.NoArgs,
	ValidArgsFunction: cobra.NoFileCompletions,
	SilenceUsage:      true,
	RunE: func(_ *cobra.Command, _ []string) error {
		if _, err := url.ParseRequestURI(config.Conf.Application.PublicURL); err != nil {
			return fmt.Errorf("invalid public URL %q: %w", config.Conf.Application.PublicURL, err)
		}

		signingMethod := auth.GetCorrectSigningMethod(config.Conf.Auth.JWT.SigningMethod)
		if signingMethod == "" {
			return fmt.Errorf("invalid or unsupported JWT signing method %q", config.Conf.Auth.JWT.SigningMethod)
		}
		config.Conf.Auth.JWT.SigningMethod = signingMethod

		if keyPath, ok := strings.CutPrefix(config.Conf.Auth.JWT.SigningKey, "file:"); ok {
			keyPath, err := filepath.Abs(keyPath)
			if err != nil {
				return fmt.Errorf("failed to get absolute path of JWT signing key file %q: %w", keyPath, err)
			}
			keyData, err := os.ReadFile(keyPath)
			if err != nil {
				return fmt.Errorf("failed to read JWT signing key file %q: %w", keyPath, err)
			}
			config.Conf.Auth.JWT.SigningKey = string(keyData)
		}

		if err := auth.LoadPrivateKey(config.Conf.Auth.JWT.SigningMethod, config.Conf.Auth.JWT.SigningKey); err != nil {
			return fmt.Errorf("failed to load private key: %w", err)
		}

		api.RegisterBuiltInProviders(config.Conf.Application.PublicURL)
		for _, provider := range config.Conf.Auth.Providers {
			if provider.ID == "" {
				return errors.New("auth provider ID cannot be empty")
			}
			if provider.GetName() == "" {
				return fmt.Errorf("unsupported auth provider %q", provider.ID)
			}
			p, err := provider.GetProvider()
			if err != nil {
				return fmt.Errorf("failed to initialize auth provider %q: %w", provider.ID, err)
			}
			goth.UseProviders(p)
		}

		if err := db.ConnectDatabase(); err != nil {
			return fmt.Errorf("database connection failed: %w", err)
		}
		defer func() {
			// Ensure the database connection is closed even if an error occurs.
			// This won't do anything if everything starts and shuts down cleanly.
			if err := db.DB.Close(); err != nil {
				log.Error("Error closing database connection", "err", err)
			}
		}()
		log.Info("Connected to database", "engine", db.DB.EngineName())

		if err := db.DB.CleanExpiredTokens(context.Background()); err != nil {
			log.Error("Error cleaning expired tokens", "err", err)
		}

		storage.CleanOverfilledStorage(context.Background())

		if err := metrics.RegisterMetrics(); err != nil {
			return fmt.Errorf("failed to register custom metrics: %w", err)
		}

		e := server.New()

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		errCh := make(chan error, 1)
		go func() {
			if err := e.Start(fmt.Sprintf(":%d", config.Conf.Application.Port)); !errors.Is(err, http.ErrServerClosed) {
				errCh <- err
			}
		}()

		select {
		case err := <-errCh:
			return fmt.Errorf("HTTP server error: %w", err)
		case <-ctx.Done():
			log.Info("Shutting down...")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := e.Shutdown(ctx); err != nil {
			return fmt.Errorf("error shutting down server: %w", err)
		}
		if err := db.DB.Close(); err != nil {
			return fmt.Errorf("error closing database connection: %w", err)
		}

		log.Info("Successfully shut down")
		return nil
	},
}

func init() {
	serverCmd.PersistentFlags().IntP("port", "p", 8080, "port on which the server will listen")
	_ = serverCmd.RegisterFlagCompletionFunc("port", cobra.NoFileCompletions)
	taurus.BindFlag("Application.Port", serverCmd.PersistentFlags().Lookup("port"))

	rootCmd.AddCommand(serverCmd)
}
