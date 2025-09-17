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
	Use:     "server",
	Aliases: []string{"start"},
	Short:   "Start the server",
	Run: func(_ *cobra.Command, _ []string) {
		if _, err := url.ParseRequestURI(config.Conf.Application.PublicURL); err != nil {
			log.Fatal("Invalid public URL", "url", config.Conf.Application.PublicURL)
		}

		signingMethod := auth.GetCorrectSigningMethod(config.Conf.Auth.JWT.SigningMethod)
		if signingMethod == "" {
			log.Fatal("Invalid or unsupported JWT signing method", "method", config.Conf.Auth.JWT.SigningMethod)
		}
		config.Conf.Auth.JWT.SigningMethod = signingMethod

		if strings.HasPrefix(config.Conf.Auth.JWT.SigningKey, "file:") {
			keyPath := strings.TrimPrefix(config.Conf.Auth.JWT.SigningKey, "file:")
			keyPath, err := filepath.Abs(keyPath)
			if err != nil {
				log.Fatal("Error getting absolute path of JWT signing key file", "err", err)
			}
			keyData, err := os.ReadFile(keyPath)
			if err != nil {
				log.Fatal("Error reading JWT signing key file", "err", err)
			}
			config.Conf.Auth.JWT.SigningKey = string(keyData)
		}

		if err := auth.LoadPrivateKey(config.Conf.Auth.JWT.SigningMethod, config.Conf.Auth.JWT.SigningKey); err != nil {
			log.Fatal("Error loading private key", "err", err)
		}

		api.RegisterBuiltInProviders(config.Conf.Application.PublicURL)
		for _, provider := range config.Conf.Auth.Providers {
			if provider.ID == "" {
				log.Fatal("OAuth provider ID cannot be empty")
			}
			if provider.GetName() == "" {
				log.Fatal("Unsupported OAuth provider", "id", provider.ID)
			}
			p, err := provider.GetProvider()
			if err != nil {
				log.Fatal("Failed to initialize OAuth provider", "id", provider.ID, "err", err)
			}
			goth.UseProviders(p)
		}

		if err := db.ConnectDatabase(); err != nil {
			log.Fatal("Error connecting to database", "err", err)
		}

		if err := db.DB.CleanExpiredTokens(context.Background()); err != nil {
			log.Error("Error cleaning expired tokens", "err", err)
		}

		storage.CleanOverfilledStorage(context.Background())

		if err := metrics.RegisterMetrics(); err != nil {
			log.Fatal("Failed to update custom metrics", "err", err)
		}

		e := server.New()

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		go func() {
			if err := e.Start(fmt.Sprintf(":%d", config.Conf.Application.Port)); errors.Is(err, http.ErrServerClosed) {
				log.Fatal("Error starting server", "err", err)
			}
		}()

		<-ctx.Done()
		log.Info("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			log.Fatal("Error shutting down server", "err", err)
		}
		if err := db.DB.Close(); err != nil {
			log.Fatal("Error closing database connection", "err", err)
		}
		log.Info("Successfully shut down")
	},
}

func init() {
	serverCmd.PersistentFlags().IntP("port", "p", 8080, "port on which the server will listen")
	_ = serverCmd.RegisterFlagCompletionFunc("port", cobra.NoFileCompletions)
	taurus.BindFlag("Application.Port", serverCmd.PersistentFlags().Lookup("port"))

	rootCmd.AddCommand(serverCmd)
}
