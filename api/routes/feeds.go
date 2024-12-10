package routes

import (
	"time"

	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gorilla/feeds"
)

func CreateFeedRoutes(router fiber.Router, basePath string, handlers ...fiber.Handler) {
	router.Get(basePath+"/feed", append(handlers, func(c *fiber.Ctx) error {
		rss, err := CreateFeed(c.BaseURL(), c.Route().Path).ToRss()
		if err != nil {
			log.Error("Error creating RSS feed", "err", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		c.Set("Content-Type", "application/rss+xml")
		return c.SendString(rss)
	})...)

	router.Get(basePath+"/feed/rss", append(handlers, func(c *fiber.Ctx) error {
		rss, err := CreateFeed(c.BaseURL(), c.Route().Path).ToRss()
		if err != nil {
			log.Error("Error creating RSS feed", "err", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		c.Set("Content-Type", "application/rss+xml")
		return c.SendString(rss)
	})...)

	router.Get(basePath+"/feed/atom", append(handlers, func(c *fiber.Ctx) error {
		atom, err := CreateFeed(c.BaseURL(), c.Route().Path).ToAtom()
		if err != nil {
			log.Error("Error creating Atom feed", "err", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		c.Set("Content-Type", "application/atom+xml")
		return c.SendString(atom)
	})...)

	router.Get(basePath+"/feed/json", append(handlers, func(c *fiber.Ctx) error {
		json, err := CreateFeed(c.BaseURL(), c.Route().Path).ToJSON()
		if err != nil {
			log.Error("Error creating JSON feed", "err", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		c.Set("Content-Type", "application/feed+json")
		return c.SendString(json)
	})...)
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
