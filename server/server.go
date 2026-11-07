package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"charm.land/log/v2"
	"github.com/goccy/go-yaml"
	openapidocs "github.com/kohkimakimoto/echo-openapidocs/v5"
	echoprometheus "github.com/labstack/echo-prometheus"
	"github.com/labstack/echo/v5"

	"github.com/libramusic/libracore"
	"github.com/libramusic/libracore/server/middleware"
	"github.com/libramusic/libracore/server/routes"
	"github.com/libramusic/libracore/server/routes/auth"
	"github.com/libramusic/libracore/sources"
	"github.com/libramusic/libracore/versions"
)

//go:generate go tool swag init -g server.go -d ./,./routes/,../media/ -o . --ot go -v3.1

//	@title			Libra API
//	@version		0.1.0-DEV
//	@description	Libra Core API providing music streaming and management capabilities.

//	@contact.name	Libra Team
//	@contact.url	https://github.com/LibraMusic/LibraCore

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

//	@servers.url	http://localhost:8080/api/v1

func New(l *libracore.Libra) *echo.Echo {
	libraSource := map[string]any{
		"id":      l.Config.Application.SourceID,
		"name":    l.Config.Application.SourceName,
		"version": versions.Libra.String(),
		"categories": []sources.Category{
			sources.CategoryMedia,
			sources.CategoryMetadata,
			sources.CategoryLyrics,
		},
		"formats": l.Config.Application.SourceFormats,
	}

	libraMeta := map[string]any{
		"version":  versions.Libra.String(),
		"database": l.MediaDB.EngineName(),
	}

	//nolint:forbidigo // Basic information on startup
	{
		fmt.Println()
		fmt.Printf("Libra v%s\n", versions.Libra.String())
		fmt.Printf("Media Database: %s\n", l.MediaDB.EngineName())
		fmt.Printf("Auth Database: %s\n", l.AuthDB.EngineName())
	}

	e := echo.New()
	e.Logger = slog.New(log.Default())
	e.JSONSerializer = &JSONSerializer{}

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set("libra", l)
			return next(c)
		}
	})

	e.Use(echoprometheus.NewMiddleware("libra"))

	e.GET("/", func(c *echo.Context) error {
		accept := c.Request().Header.Get(echo.HeaderAccept)
		switch accept {
		case echo.MIMEApplicationJSON:
			return c.JSON(http.StatusOK, &libraSource)
		case echo.MIMETextHTML:
			// TODO: Implement.
			return c.HTML(http.StatusOK, "<h1>Libra</h1>")
		}

		return c.NoContent(http.StatusNotAcceptable)
	})

	e.GET("/source", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, &libraSource)
	})

	e.GET("/meta", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, &libraMeta)
	})

	e.GET("/app", func(c *echo.Context) error {
		// TODO: Implement.
		return c.HTML(http.StatusOK, "<h1>Libra</h1>")
	})

	e.GET("/metrics", echoprometheus.NewHandler())

	apiGroup := e.Group("/api")

	authGroup := apiGroup.Group("/auth")
	authGroup.POST("/register", auth.Register)
	authGroup.POST("/login", auth.Login)
	authGroup.POST("/login/:provider", auth.LoginProvider)
	authGroup.POST("/logout", auth.Logout, middleware.JWTProtected(l))
	authGroup.GET("/connect/:provider", auth.ConnectProvider, middleware.JWTProtected(l))
	authGroup.GET("/callback/:provider", auth.ProviderCallback)
	authGroup.POST("/disconnect/:provider", auth.DisconnectProvider, middleware.JWTProtected(l))

	v1Group := apiGroup.Group("/v1")

	openapidocs.ElementsDocuments(e, "/api/v1/docs", openapidocs.ElementsConfig{
		SpecUrl: "/api/v1/openapi.json",
		Title:   "Libra API",
	})

	v1Group.GET("/playables", routes.V1Playables)
	routes.CreateFeedRoutes(v1Group, "/playables", "{} feed for all playables")
	v1Group.GET("/playables/:id", routes.V1UserPlayables)
	routes.CreateFeedRoutes(v1Group, "/playables/:id", "{} feed for user's playables")
	v1Group.GET("/search", routes.V1Search, middleware.GlobalJWTProtected(l))

	// START TO REFRACTOR
	v1Group.GET("/track/:id", routes.V1Track, middleware.GlobalJWTProtected(l))
	v1Group.GET("/track/:id/is_stored", routes.V1TrackIsStored, middleware.GlobalJWTProtected(l))
	v1Group.GET("/track/:id/stream", routes.V1TrackStream, middleware.GlobalJWTProtected(l))
	v1Group.GET("/track/:id/cover", routes.V1TrackCover, middleware.GlobalJWTProtected(l))
	v1Group.GET("/track/:id/lyrics", routes.V1TrackLyrics, middleware.GlobalJWTProtected(l))
	v1Group.GET("/track/:id/lyrics/:lang", routes.V1TrackLyricsLang, middleware.GlobalJWTProtected(l))

	v1Group.GET("/album/:id", routes.V1Album, middleware.GlobalJWTProtected(l))
	v1Group.GET("/album/:id/cover", routes.V1AlbumCover, middleware.GlobalJWTProtected(l))
	v1Group.GET("/album/:id/tracks", routes.V1AlbumTracks, middleware.GlobalJWTProtected(l))

	v1Group.GET("/video/:id", routes.V1Video, middleware.GlobalJWTProtected(l))
	v1Group.GET("/video/:id/is_stored", routes.V1VideoIsStored, middleware.GlobalJWTProtected(l))
	v1Group.GET("/video/:id/stream", routes.V1VideoStream, middleware.GlobalJWTProtected(l))
	v1Group.GET("/video/:id/cover", routes.V1VideoCover, middleware.GlobalJWTProtected(l))
	v1Group.GET("/video/:id/subtitles", routes.V1VideoSubtitles, middleware.GlobalJWTProtected(l))
	v1Group.GET("/video/:id/subtitles/:lang", routes.V1VideoSubtitlesLang, middleware.GlobalJWTProtected(l))
	v1Group.GET("/video/:id/lyrics", routes.V1VideoSubtitles, middleware.GlobalJWTProtected(l))
	v1Group.GET("/video/:id/lyrics/:lang", routes.V1VideoSubtitlesLang, middleware.GlobalJWTProtected(l))

	v1Group.GET("/playlist/:id", routes.V1Playlist, middleware.GlobalJWTProtected(l))
	v1Group.GET("/playlist/:id/cover", routes.V1PlaylistCover, middleware.GlobalJWTProtected(l))
	v1Group.GET("/playlist/:id/tracks", routes.V1PlaylistTracks, middleware.GlobalJWTProtected(l))

	v1Group.GET("/artist/:id", routes.V1Artist, middleware.GlobalJWTProtected(l))
	v1Group.GET("/artist/:id/cover", routes.V1ArtistCover, middleware.GlobalJWTProtected(l))
	v1Group.GET("/artist/:id/albums", routes.V1ArtistAlbums, middleware.GlobalJWTProtected(l))
	v1Group.GET("/artist/:id/tracks", routes.V1ArtistTracks, middleware.GlobalJWTProtected(l))
	// END TO REFRACTOR

	v1Spec := OpenAPISpec(l)
	v1SpecYAML, err := yaml.Marshal(v1Spec)
	if err != nil {
		log.Fatal("Error marshalling OpenAPI spec to YAML", "err", err)
	}

	v1Group.GET("/openapi.json", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, v1Spec)
	})

	v1Group.GET("/openapi.yaml", func(c *echo.Context) error {
		return c.String(http.StatusOK, string(v1SpecYAML))
	})

	return e
}
