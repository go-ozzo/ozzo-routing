package routing_test

import (
	"github.com/go-ozzo/ozzo-routing"
	"github.com/go-ozzo/ozzo-routing/access"
	"github.com/go-ozzo/ozzo-routing/content"
	"github.com/go-ozzo/ozzo-routing/fault"
	"github.com/go-ozzo/ozzo-routing/file"
	"github.com/go-ozzo/ozzo-routing/slash"
	"log"
	"net/http"
)

func Example() {
	router := routing.New()

	router.Use(
		// all these handlers are shared by every route
		access.Logger(log.Printf),
		slash.Remover(http.StatusMovedPermanently),
		fault.Recovery(log.Printf),
	)

	// serve RESTful APIs
	api := router.Group("/api")
	api.Use(
		// these handlers are shared by the routes in the api group only
		content.TypeNegotiator(content.JSON, content.XML),
	)
	api.Get("/users", func(c *routing.Context) error {
		return c.Write("user list")
	})
	api.Post("/users", func(c *routing.Context) error {
		return c.Write("create a new user")
	})
	api.Put(`/users/<id:\d+>`, func(c *routing.Context) error {
		return c.Write("update user " + c.Param("id"))
	})

	// serve index file
	router.Get("/", file.Content("ui/index.html"))
	// serve files under the "ui" subdirectory
	router.Get("/*", file.Server(file.PathMap{
		"/": "/ui/",
	}))

	http.Handle("/", router)
	http.ListenAndServe(":8080", nil)
}
