package cmds

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/goccy/go-yaml"
	openapidocs "github.com/kohkimakimoto/echo-openapidocs"
	"github.com/labstack/echo/v4"
	"github.com/libramusic/taurus"
	"github.com/markbates/goth"
	"github.com/spf13/cobra"

	"github.com/libramusic/libracore/api"
	"github.com/libramusic/libracore/api/middleware"
	"github.com/libramusic/libracore/api/routes"
	"github.com/libramusic/libracore/config"
	"github.com/libramusic/libracore/db"
	"github.com/libramusic/libracore/sources"
	"github.com/libramusic/libracore/storage"
	"github.com/libramusic/libracore/utils"
)

var serverCmd = &cobra.Command{
	Use:     "server",
	Aliases: []string{"start"},
	Short:   "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		utils.SetupLogger(config.Conf.Logs.LogFormat, config.Conf.Logs.LogLevel)

		signingMethod := utils.GetCorrectSigningMethod(config.Conf.Auth.JWT.SigningMethod)
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

		if err := utils.LoadPrivateKey(config.Conf.Auth.JWT.SigningMethod, config.Conf.Auth.JWT.SigningKey); err != nil {
			log.Fatal("Error loading private key", "err", err)
		}

		api.InitProviderFactories(config.Conf.Application.PublicURL)
		for _, provider := range config.Conf.Auth.OAuth.Providers {
			if provider.ID == "" {
				log.Fatal("OAuth provider ID cannot be empty")
			}
			if provider.GetName() == "" {
				log.Fatal("Unsupported OAuth provider", "id", provider.ID)
			}
			if p, err := provider.GetProvider(); err != nil {
				log.Fatal("Failed to initialize OAuth provider", "id", provider.ID, "err", err)
			} else {
				goth.UseProviders(p)
			}
		}

		if err := db.ConnectDatabase(); err != nil {
			log.Fatal("Error connecting to database", "err", err)
		}

		if err := db.DB.CleanExpiredTokens(); err != nil {
			log.Error("Error cleaning expired tokens", "err", err)
		}

		storage.CleanOverfilledStorage()

		sources.InitManager()

		// Test code below (TODO: Remove)
		s, err := sources.InitYouTubeSource()
		if err != nil {
			log.Fatal("Error initializing YouTube source", "err", err)
		}
		a, b := s.Search("Lord of Ashes", 5, 0, map[string]interface{}{})
		fmt.Println(a)
		fmt.Println(b)
		// fmt.Println(s.ContainsURL("https://www.youtube.com/watch?v=orimodrogvd"))
		// fmt.Println(s.ContainsURL("https://www.youtube.com/watch?v=uGxcco8Uq6A"))
		// Test code above (TODO: Remove)

		libraService := echo.Map{
			"id":           config.Conf.Application.SourceID,
			"name":         config.Conf.Application.SourceName,
			"version":      utils.LibraVersion.String(),
			"source_types": []string{"content", "metadata", "lyrics"},
			"media_types":  config.Conf.Application.MediaTypes,
		}

		libraMeta := echo.Map{
			"version":  utils.LibraVersion.String(),
			"database": db.DB.EngineName(),
		}

		fmt.Println()
		fmt.Printf("Libra v%s\n", utils.LibraVersion.String())
		fmt.Printf("Database: %s\n", db.DB.EngineName())

		v1Spec := api.V1OpenAPI3Spec()
		v1SpecYAML, err := yaml.Marshal(v1Spec)
		if err != nil {
			log.Fatal("Error marshalling OpenAPI spec to YAML", "err", err)
		}

		e := echo.New()
		e.JSONSerializer = &api.GoJSONSerializer{}

		e.GET("/", func(c echo.Context) error {
			accept := c.Request().Header.Get(echo.HeaderAccept)
			if accept == echo.MIMEApplicationJSON {
				return c.JSON(http.StatusOK, &libraService)
			} else if accept == echo.MIMETextHTML {
				// TODO: Implement
				return c.HTML(http.StatusOK, "<h1>Libra</h1>")
			}

			return c.NoContent(http.StatusNotAcceptable)
		})

		e.GET("/meta", func(c echo.Context) error {
			return c.JSON(http.StatusOK, &libraMeta)
		})

		e.GET("/service", func(c echo.Context) error {
			return c.JSON(http.StatusOK, &libraService)
		})

		e.GET("/app", func(c echo.Context) error {
			// TODO: Implement
			return c.HTML(http.StatusOK, "<h1>Libra</h1>")
		})

		api := e.Group("/api")

		auth := api.Group("/auth")
		auth.POST("/register", routes.Register)
		auth.POST("/login", routes.Login)
		auth.POST("/logout", routes.Logout, middleware.JWTProtected)

		v1 := api.Group("/v1")

		openapidocs.ElementsDocuments(e, "/api/v1/docs", openapidocs.ElementsConfig{
			SpecUrl: "/api/v1/openapi.json",
			Title:   "Libra API",
		})

		v1.GET("/playables", routes.V1Playables)
		routes.CreateFeedRoutes(v1, "/playables")
		v1.GET("/search", routes.V1Search, middleware.GlobalJWTProtected)

		// START TO REFRACTOR
		v1.GET("/track/:id", routes.V1Track, middleware.GlobalJWTProtected)
		v1.GET("/track/:id/is_stored", routes.V1TrackIsStored, middleware.GlobalJWTProtected)
		v1.GET("/track/:id/stream", routes.V1TrackStream, middleware.GlobalJWTProtected)
		v1.GET("/track/:id/cover", routes.V1TrackCover, middleware.GlobalJWTProtected)
		v1.GET("/track/:id/lyrics", routes.V1TrackLyrics, middleware.GlobalJWTProtected)
		v1.GET("/track/:id/lyrics/:lang", routes.V1TrackLyricsLang, middleware.GlobalJWTProtected)

		v1.GET("/album/:id", routes.V1Album, middleware.GlobalJWTProtected)
		v1.GET("/album/:id/cover", routes.V1AlbumCover, middleware.GlobalJWTProtected)
		v1.GET("/album/:id/tracks", routes.V1AlbumTracks, middleware.GlobalJWTProtected)

		v1.GET("/video/:id", routes.V1Video, middleware.GlobalJWTProtected)
		v1.GET("/video/:id/is_stored", routes.V1VideoIsStored, middleware.GlobalJWTProtected)
		v1.GET("/video/:id/stream", routes.V1VideoStream, middleware.GlobalJWTProtected)
		v1.GET("/video/:id/cover", routes.V1VideoCover, middleware.GlobalJWTProtected)
		v1.GET("/video/:id/subtitles", routes.V1VideoSubtitles, middleware.GlobalJWTProtected)
		v1.GET("/video/:id/subtitles/:lang", routes.V1VideoSubtitlesLang, middleware.GlobalJWTProtected)
		v1.GET("/video/:id/lyrics", routes.V1VideoSubtitles, middleware.GlobalJWTProtected)
		v1.GET("/video/:id/lyrics/:lang", routes.V1VideoSubtitlesLang, middleware.GlobalJWTProtected)

		v1.GET("/playlist/:id", routes.V1Playlist, middleware.GlobalJWTProtected)
		v1.GET("/playlist/:id/cover", routes.V1PlaylistCover, middleware.GlobalJWTProtected)
		v1.GET("/playlist/:id/tracks", routes.V1PlaylistTracks, middleware.GlobalJWTProtected)

		v1.GET("/artist/:id", routes.V1Artist, middleware.GlobalJWTProtected)
		v1.GET("/artist/:id/cover", routes.V1ArtistCover, middleware.GlobalJWTProtected)
		v1.GET("/artist/:id/albums", routes.V1ArtistAlbums, middleware.GlobalJWTProtected)
		v1.GET("/artist/:id/tracks", routes.V1ArtistTracks, middleware.GlobalJWTProtected)
		// END TO REFRACTOR

		v1.GET("/openapi.json", func(c echo.Context) error {
			return c.JSON(http.StatusOK, v1Spec)
		})

		v1.GET("/openapi.yaml", func(c echo.Context) error {
			return c.String(http.StatusOK, string(v1SpecYAML))
		})

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		go func() {
			if err := e.Start(fmt.Sprintf(":%d", config.Conf.Application.Port)); err != nil && err != http.ErrServerClosed {
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
