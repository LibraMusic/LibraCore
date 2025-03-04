package server

import (
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/goccy/go-yaml"
	openapidocs "github.com/kohkimakimoto/echo-openapidocs"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"

	"github.com/libramusic/libracore/api"
	"github.com/libramusic/libracore/api/middleware"
	"github.com/libramusic/libracore/api/routes"
	"github.com/libramusic/libracore/config"
	"github.com/libramusic/libracore/db"
	"github.com/libramusic/libracore/utils"
)

//go:generate go tool swag init -g server.go -d ./,../routes/,../../types/ -o . --ot go -v3.1

//	@title			Libra API
//	@version		0.1.0-DEV
//	@description	Libra Core API providing music streaming and management capabilities.

//	@contact.name	Libra Team
//	@contact.url	https://github.com/LibraMusic/LibraCore

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

//	@servers.url	http://localhost:8080/api/v1

func InitServer() *echo.Echo {
	libraSource := echo.Map{
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

	//nolint:forbidigo // Basic information on startup
	{
		fmt.Println()
		fmt.Printf("Libra v%s\n", utils.LibraVersion.String())
		fmt.Printf("Database: %s\n", db.DB.EngineName())
	}

	e := echo.New()
	e.JSONSerializer = &api.GoJSONSerializer{}

	e.Use(echoprometheus.NewMiddleware("libra"))

	e.GET("/", func(c echo.Context) error {
		accept := c.Request().Header.Get(echo.HeaderAccept)
		if accept == echo.MIMEApplicationJSON {
			return c.JSON(http.StatusOK, &libraSource)
		} else if accept == echo.MIMETextHTML {
			// TODO: Implement.
			return c.HTML(http.StatusOK, "<h1>Libra</h1>")
		}

		return c.NoContent(http.StatusNotAcceptable)
	})

	e.GET("/source", func(c echo.Context) error {
		return c.JSON(http.StatusOK, &libraSource)
	})

	e.GET("/meta", func(c echo.Context) error {
		return c.JSON(http.StatusOK, &libraMeta)
	})

	e.GET("/app", func(c echo.Context) error {
		// TODO: Implement.
		return c.HTML(http.StatusOK, "<h1>Libra</h1>")
	})

	e.GET("/metrics", echoprometheus.NewHandler())

	apiGroup := e.Group("/api")

	authGroup := apiGroup.Group("/auth")
	authGroup.POST("/register", routes.Register)
	authGroup.POST("/login", routes.Login)
	authGroup.POST("/logout", routes.Logout, middleware.JWTProtected)
	authGroup.POST("/logout/:provider", routes.OAuthLogout)
	authGroup.GET("/:provider", routes.OAuth)
	authGroup.GET("/:provider/callback", routes.OAuthCallback)

	v1Group := apiGroup.Group("/v1")

	openapidocs.ElementsDocuments(e, "/api/v1/docs", openapidocs.ElementsConfig{
		SpecUrl: "/api/v1/openapi.json",
		Title:   "Libra API",
	})

	v1Group.GET("/playables", routes.V1Playables)
	routes.CreateFeedRoutes(v1Group, "/playables", "{} feed for all playables")
	v1Group.GET("/playables/:id", routes.V1UserPlayables)
	routes.CreateFeedRoutes(v1Group, "/playables/:id", "{} feed for user's playables")
	v1Group.GET("/search", routes.V1Search, middleware.GlobalJWTProtected)

	// START TO REFRACTOR
	v1Group.GET("/track/:id", routes.V1Track, middleware.GlobalJWTProtected)
	v1Group.GET("/track/:id/is_stored", routes.V1TrackIsStored, middleware.GlobalJWTProtected)
	v1Group.GET("/track/:id/stream", routes.V1TrackStream, middleware.GlobalJWTProtected)
	v1Group.GET("/track/:id/cover", routes.V1TrackCover, middleware.GlobalJWTProtected)
	v1Group.GET("/track/:id/lyrics", routes.V1TrackLyrics, middleware.GlobalJWTProtected)
	v1Group.GET("/track/:id/lyrics/:lang", routes.V1TrackLyricsLang, middleware.GlobalJWTProtected)

	v1Group.GET("/album/:id", routes.V1Album, middleware.GlobalJWTProtected)
	v1Group.GET("/album/:id/cover", routes.V1AlbumCover, middleware.GlobalJWTProtected)
	v1Group.GET("/album/:id/tracks", routes.V1AlbumTracks, middleware.GlobalJWTProtected)

	v1Group.GET("/video/:id", routes.V1Video, middleware.GlobalJWTProtected)
	v1Group.GET("/video/:id/is_stored", routes.V1VideoIsStored, middleware.GlobalJWTProtected)
	v1Group.GET("/video/:id/stream", routes.V1VideoStream, middleware.GlobalJWTProtected)
	v1Group.GET("/video/:id/cover", routes.V1VideoCover, middleware.GlobalJWTProtected)
	v1Group.GET("/video/:id/subtitles", routes.V1VideoSubtitles, middleware.GlobalJWTProtected)
	v1Group.GET("/video/:id/subtitles/:lang", routes.V1VideoSubtitlesLang, middleware.GlobalJWTProtected)
	v1Group.GET("/video/:id/lyrics", routes.V1VideoSubtitles, middleware.GlobalJWTProtected)
	v1Group.GET("/video/:id/lyrics/:lang", routes.V1VideoSubtitlesLang, middleware.GlobalJWTProtected)

	v1Group.GET("/playlist/:id", routes.V1Playlist, middleware.GlobalJWTProtected)
	v1Group.GET("/playlist/:id/cover", routes.V1PlaylistCover, middleware.GlobalJWTProtected)
	v1Group.GET("/playlist/:id/tracks", routes.V1PlaylistTracks, middleware.GlobalJWTProtected)

	v1Group.GET("/artist/:id", routes.V1Artist, middleware.GlobalJWTProtected)
	v1Group.GET("/artist/:id/cover", routes.V1ArtistCover, middleware.GlobalJWTProtected)
	v1Group.GET("/artist/:id/albums", routes.V1ArtistAlbums, middleware.GlobalJWTProtected)
	v1Group.GET("/artist/:id/tracks", routes.V1ArtistTracks, middleware.GlobalJWTProtected)
	// END TO REFRACTOR

	v1Spec := GetOpenAPISpec()
	v1SpecYAML, err := yaml.Marshal(v1Spec)
	if err != nil {
		log.Fatal("Error marshalling OpenAPI spec to YAML", "err", err)
	}

	v1Group.GET("/openapi.json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, v1Spec)
	})

	v1Group.GET("/openapi.yaml", func(c echo.Context) error {
		return c.String(http.StatusOK, string(v1SpecYAML))
	})

	return e
}
