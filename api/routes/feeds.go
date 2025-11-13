package routes

import (
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gorilla/feeds"
	"github.com/labstack/echo/v4"
)

// FeedRouteDoc definition for dynamic OpenAPI documentation.
type FeedRouteDoc struct {
	BasePath string
	Path     string
	Summary  string
	Type     string
}

var FeedRoutesDoc []FeedRouteDoc

func ConvertPathFormat(path string) string {
	// convert from echo format to openapi format
	// e.g. /path/:param -> /path/{param}
	result := ""
	parts := strings.Split(path, "/")
	var resultSb28 strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}

		split := strings.Split(part, ":")
		if len(split) > 1 {
			resultSb28.WriteString("/" + "{" + split[1] + "}")
			var resultSb36 strings.Builder
			for i := 2; i < len(split); i++ {
				resultSb36.WriteString(split[i])
			}
			result += resultSb36.String()
		} else {
			resultSb28.WriteString("/" + part)
		}
	}
	result += resultSb28.String()
	return result
}

func CreateFeedRoutes(e *echo.Group, basePath, baseSummary string, handlers ...echo.MiddlewareFunc) {
	// Define a helper to add route and docs.
	addRoute := func(path, feedType string, h echo.HandlerFunc) {
		fullPath := basePath + path
		FeedRoutesDoc = append(FeedRoutesDoc, FeedRouteDoc{
			BasePath: ConvertPathFormat(basePath),
			Path:     ConvertPathFormat(fullPath),
			Summary:  strings.ReplaceAll(baseSummary, "{}", feedType),
			Type:     feedType,
		})
		e.GET(fullPath, h, handlers...)
	}

	addRoute("/feed", "RSS", func(c echo.Context) error {
		rss, err := CreateFeed(c.Scheme()+"://"+c.Request().Host, c.Path()).ToRss()
		if err != nil {
			log.Error("Error creating RSS feed", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.XMLBlob(http.StatusOK, []byte(rss))
	})

	addRoute("/feed/rss", "RSS", func(c echo.Context) error {
		rss, err := CreateFeed(c.Scheme()+"://"+c.Request().Host, c.Path()).ToRss()
		if err != nil {
			log.Error("Error creating RSS feed", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.XMLBlob(http.StatusOK, []byte(rss))
	})

	addRoute("/feed/atom", "Atom", func(c echo.Context) error {
		atom, err := CreateFeed(c.Scheme()+"://"+c.Request().Host, c.Path()).ToAtom()
		if err != nil {
			log.Error("Error creating Atom feed", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.XMLBlob(http.StatusOK, []byte(atom))
	})

	addRoute("/feed/json", "JSON", func(c echo.Context) error {
		json, err := CreateFeed(c.Scheme()+"://"+c.Request().Host, c.Path()).ToJSON()
		if err != nil {
			log.Error("Error creating JSON feed", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.JSONBlob(http.StatusOK, []byte(json))
	})
}

func CreateFeed(baseURL, routePath string) *feeds.Feed {
	// TODO: Get data from route at basePath for the resulting value here.
	feed := &feeds.Feed{
		Title:       "Libra",
		Link:        &feeds.Link{Href: baseURL + routePath},
		Description: "A modern, open-source music server and streaming platform",
		Author:      &feeds.Author{Name: "Libra Team"},
		Created:     time.Now(),
		Updated:     time.Now(),
	}

	feed.Items = []*feeds.Item{
		{
			Title:       "Test",
			Description: "A test thingy",
			Author:      &feeds.Author{Name: "DevReaper0", Email: "devreaper0@gmail.com"},
			Created:     time.Now(),
		},
		{
			Title:       "Test B",
			Description: "Another test thingy",
			Author:      &feeds.Author{Name: "DevReaper0", Email: "devreaper0@gmail.com"},
			Created:     time.Now(),
		},
	}

	return feed
}
