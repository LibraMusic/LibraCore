package routes

import (
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gorilla/feeds"
	"github.com/labstack/echo/v4"
)

func CreateFeedRoutes(e *echo.Group, basePath string, handlers ...echo.MiddlewareFunc) {
	e.GET(basePath+"/feed", func(c echo.Context) error {
		rss, err := CreateFeed(c.Scheme()+"://"+c.Request().Host, c.Path()).ToRss()
		if err != nil {
			log.Error("Error creating RSS feed", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.XMLBlob(http.StatusOK, []byte(rss))
	}, handlers...)

	e.GET(basePath+"/feed/rss", func(c echo.Context) error {
		rss, err := CreateFeed(c.Scheme()+"://"+c.Request().Host, c.Path()).ToRss()
		if err != nil {
			log.Error("Error creating RSS feed", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.XMLBlob(http.StatusOK, []byte(rss))
	}, handlers...)

	e.GET(basePath+"/feed/atom", func(c echo.Context) error {
		atom, err := CreateFeed(c.Scheme()+"://"+c.Request().Host, c.Path()).ToAtom()
		if err != nil {
			log.Error("Error creating Atom feed", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.XMLBlob(http.StatusOK, []byte(atom))
	}, handlers...)

	e.GET(basePath+"/feed/json", func(c echo.Context) error {
		json, err := CreateFeed(c.Scheme()+"://"+c.Request().Host, c.Path()).ToJSON()
		if err != nil {
			log.Error("Error creating JSON feed", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.JSONBlob(http.StatusOK, []byte(json))
	}, handlers...)
}

func CreateFeed(baseURL string, routePath string) *feeds.Feed {
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
