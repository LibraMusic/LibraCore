package cmds

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"

	"github.com/LibraMusic/LibraCore/api"
	"github.com/LibraMusic/LibraCore/api/middleware"
	"github.com/LibraMusic/LibraCore/api/routes"
	"github.com/LibraMusic/LibraCore/config"
	"github.com/LibraMusic/LibraCore/db"
	"github.com/LibraMusic/LibraCore/sources"
	"github.com/LibraMusic/LibraCore/storage"
	"github.com/LibraMusic/LibraCore/utils"
)

var serverCmd = &cobra.Command{
	Use:     "server",
	Aliases: []string{"start"},
	Short:   "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		utils.SetupLogger(config.Conf.Logs.LogFormat, config.Conf.Logs.LogLevel)

		signingMethod := utils.GetCorrectSigningMethod(config.Conf.Auth.JWTSigningMethod)
		if signingMethod == "" {
			log.Fatal("Invalid or unsupported JWT signing method", "method", config.Conf.Auth.JWTSigningMethod)
		}
		config.Conf.Auth.JWTSigningMethod = signingMethod

		if strings.HasPrefix(config.Conf.Auth.JWTSigningKey, "file:") {
			keyPath := strings.TrimPrefix(config.Conf.Auth.JWTSigningKey, "file:")
			keyPath, err := filepath.Abs(keyPath)
			if err != nil {
				log.Fatal("Error getting absolute path of JWT signing key file", "err", err)
			}
			keyData, err := os.ReadFile(keyPath)
			if err != nil {
				log.Fatal("Error reading JWT signing key file", "err", err)
			}
			config.Conf.Auth.JWTSigningKey = string(keyData)
		}

		err := utils.LoadPrivateKey(config.Conf.Auth.JWTSigningMethod, config.Conf.Auth.JWTSigningKey)
		if err != nil {
			log.Fatal("Error loading private key", "err", err)
		}

		err = db.ConnectDatabase()
		if err != nil {
			log.Fatal("Error connecting to database", "err", err)
		}

		err = db.DB.CleanExpiredTokens()
		if err != nil {
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
		//fmt.Println(s.ContainsURL("https://www.youtube.com/watch?v=orimodrogvd"))
		//fmt.Println(s.ContainsURL("https://www.youtube.com/watch?v=uGxcco8Uq6A"))
		// Test code above (TODO: Remove)

		libraService := fiber.Map{
			"id":           config.Conf.Application.SourceID,
			"name":         config.Conf.Application.SourceName,
			"version":      utils.LibraVersion.String(),
			"source_types": []string{"content", "metadata", "lyrics"},
			"media_types":  config.Conf.Application.MediaTypes,
		}

		libraMeta := fiber.Map{
			"version":  utils.LibraVersion.String(),
			"database": db.DB.EngineName(),
		}

		fmt.Println()
		fmt.Printf("Libra v%s\n", utils.LibraVersion.String())
		fmt.Printf("Database: %s\n", db.DB.EngineName())

		v1Spec := api.V1OpenAPI3Spec()
		v1SpecJSON, err := json.Marshal(v1Spec)
		if err != nil {
			log.Fatal("Error marshalling OpenAPI spec to JSON", "err", err)
		}
		v1SpecYAML, err := yaml.Marshal(v1Spec)
		if err != nil {
			log.Fatal("Error marshalling OpenAPI spec to YAML", "err", err)
		}

		app := fiber.New(fiber.Config{
			JSONEncoder: json.Marshal,
			JSONDecoder: json.Unmarshal,
		})

		app.Get("/", func(c *fiber.Ctx) error {
			offer := c.Accepts(fiber.MIMEApplicationJSON, fiber.MIMETextHTML)
			if offer == fiber.MIMEApplicationJSON {
				return c.JSON(&libraService)
			} else if offer == fiber.MIMETextHTML {
				// TODO: Implement
				return c.SendString("<h1>Libra</h1>")
			}

			return c.SendStatus(fiber.StatusNotAcceptable)
		})

		app.Get("/meta", func(c *fiber.Ctx) error {
			return c.JSON(&libraMeta)
		})

		app.Get("/service", func(c *fiber.Ctx) error {
			return c.JSON(&libraService)
		})

		app.Get("/app", func(c *fiber.Ctx) error {
			// TODO: Implement
			return c.SendString("<h1>Libra</h1>")
		})

		api := app.Group("/api")

		auth := api.Group("/auth")
		auth.Post("/register", routes.Register)
		auth.Post("/login", routes.Login)
		auth.Post("/logout", middleware.JWTProtected, routes.Logout)

		v1 := api.Group("/v1")

		v1.Use(swagger.New(swagger.Config{
			BasePath:    "/api/v1/",
			FilePath:    "/openapi.json",
			FileContent: v1SpecJSON,
			Path:        "docs",
			Title:       "Libra API",
		}))

		v1.Get("/playables", routes.V1Playables)
		routes.CreateFeedRoutes(v1, "/playables")
		v1.Get("/search", middleware.GlobalJWTProtected, routes.V1Search)

		// START TO REFRACTOR
		v1.Get("/track/:id", middleware.GlobalJWTProtected, routes.V1Track)
		v1.Get("/track/:id/is_stored", middleware.GlobalJWTProtected, routes.V1TrackIsStored)
		v1.Get("/track/:id/stream", middleware.GlobalJWTProtected, routes.V1TrackStream)
		v1.Get("/track/:id/cover", middleware.GlobalJWTProtected, routes.V1TrackCover)
		v1.Get("/track/:id/lyrics", middleware.GlobalJWTProtected, routes.V1TrackLyrics)
		v1.Get("/track/:id/lyrics/:lang", middleware.GlobalJWTProtected, routes.V1TrackLyricsLang)

		v1.Get("/album/:id", middleware.GlobalJWTProtected, routes.V1Album)
		v1.Get("/album/:id/cover", middleware.GlobalJWTProtected, routes.V1AlbumCover)
		v1.Get("/album/:id/tracks", middleware.GlobalJWTProtected, routes.V1AlbumTracks)

		v1.Get("/video/:id", middleware.GlobalJWTProtected, routes.V1Video)
		v1.Get("/video/:id/is_stored", middleware.GlobalJWTProtected, routes.V1VideoIsStored)
		v1.Get("/video/:id/stream", middleware.GlobalJWTProtected, routes.V1VideoStream)
		v1.Get("/video/:id/cover", middleware.GlobalJWTProtected, routes.V1VideoCover)
		v1.Get("/video/:id/subtitles", middleware.GlobalJWTProtected, routes.V1VideoSubtitles)
		v1.Get("/video/:id/subtitles/:lang", middleware.GlobalJWTProtected, routes.V1VideoSubtitlesLang)
		v1.Get("/video/:id/lyrics", middleware.GlobalJWTProtected, routes.V1VideoSubtitles)
		v1.Get("/video/:id/lyrics/:lang", middleware.GlobalJWTProtected, routes.V1VideoSubtitlesLang)

		v1.Get("/playlist/:id", middleware.GlobalJWTProtected, routes.V1Playlist)
		v1.Get("/playlist/:id/cover", middleware.GlobalJWTProtected, routes.V1PlaylistCover)
		v1.Get("/playlist/:id/tracks", middleware.GlobalJWTProtected, routes.V1PlaylistTracks)

		v1.Get("/artist/:id", middleware.GlobalJWTProtected, routes.V1Artist)
		v1.Get("/artist/:id/cover", middleware.GlobalJWTProtected, routes.V1ArtistCover)
		v1.Get("/artist/:id/albums", middleware.GlobalJWTProtected, routes.V1ArtistAlbums)
		v1.Get("/artist/:id/tracks", middleware.GlobalJWTProtected, routes.V1ArtistTracks)
		// END TO REFRACTOR

		v1.Get("/openapi.json", func(c *fiber.Ctx) error {
			return c.JSON(v1SpecJSON)
		})

		v1.Get("/openapi.yaml", func(c *fiber.Ctx) error {
			return c.Send(v1SpecYAML)
		})

		if err = app.Listen(fmt.Sprintf(":%d", config.Conf.Application.Port)); err != nil {
			log.Fatal("Error starting server", "err", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
