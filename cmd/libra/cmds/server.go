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

	"charm.land/log/v2"
	"github.com/labstack/echo/v5"
	"github.com/markbates/goth"
	"github.com/spf13/cobra"

	"github.com/libramusic/taurus/v2"

	"github.com/libramusic/libracore/server"
	"github.com/libramusic/libracore/server/metrics"
	"github.com/libramusic/libracore/server/routes/auth"
	"github.com/libramusic/libracore/server/routes/auth/providers"
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
		if _, err := url.ParseRequestURI(l.Config.Application.PublicURL); err != nil {
			return fmt.Errorf("invalid public URL %q: %w", l.Config.Application.PublicURL, err)
		}

		signingMethod := auth.CorrectSigningMethod(l.Config.Auth.JWT.SigningMethod)
		if signingMethod == "" {
			return fmt.Errorf("invalid or unsupported JWT signing method %q", l.Config.Auth.JWT.SigningMethod)
		}
		l.Config.Auth.JWT.SigningMethod = signingMethod

		if keyPath, ok := strings.CutPrefix(l.Config.Auth.JWT.SigningKey, "file:"); ok {
			keyPath, err := filepath.Abs(keyPath)
			if err != nil {
				return fmt.Errorf("failed to get absolute path of JWT signing key file %q: %w", keyPath, err)
			}
			keyData, err := os.ReadFile(keyPath)
			if err != nil {
				return fmt.Errorf("failed to read JWT signing key file %q: %w", keyPath, err)
			}
			l.Config.Auth.JWT.SigningKey = string(keyData)
		}

		if err := auth.LoadPrivateKey(l.Config.Auth.JWT.SigningMethod, l.Config.Auth.JWT.SigningKey); err != nil {
			return fmt.Errorf("failed to load private key: %w", err)
		}

		providers.RegisterBuiltInProviders(l.Config.Application.PublicURL)
		for _, provider := range l.Config.Auth.Providers {
			if provider.ID == "" {
				return errors.New("auth provider ID cannot be empty")
			}
			if provider.Name() == "" {
				return fmt.Errorf("unsupported auth provider %q", provider.ID)
			}
			p, err := provider.GothProvider()
			if err != nil {
				return fmt.Errorf("failed to initialize auth provider %q: %w", provider.ID, err)
			}
			goth.UseProviders(p)
		}

		if err := l.Connect(); err != nil {
			return err
		}
		defer func() {
			// Ensure the instance is closed even if an error occurs.
			// This won't do anything if everything starts and shuts down cleanly.
			if err := l.Close(); err != nil {
				log.Error("Error shutting down", "err", err)
			}
		}()
		log.Info("Connected to databases")

		if err := l.AuthDB.CleanExpiredTokens(context.Background()); err != nil {
			log.Error("Error cleaning expired tokens", "err", err)
		}

		storage.CleanOverfilledStorage(context.Background(), l.Config.Storage, l.MediaDB)

		if err := metrics.RegisterMetrics(l); err != nil {
			return fmt.Errorf("failed to register custom metrics: %w", err)
		}

		e := server.New(l)

		sc := echo.StartConfig{
			Address:    fmt.Sprintf(":%d", l.Config.Application.Port),
			HideBanner: true,
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		errCh := make(chan error, 1)
		go func() {
			if err := sc.Start(ctx, e); !errors.Is(err, http.ErrServerClosed) {
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

		if err := l.Close(); err != nil {
			return fmt.Errorf("error shutting down: %w", err)
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
