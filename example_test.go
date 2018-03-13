package routing_test

import (
	"context"
	"github.com/ltick/tick-routing"
	"github.com/ltick/tick-routing/access"
	"github.com/ltick/tick-routing/content"
	"github.com/ltick/tick-routing/fault"
	"github.com/ltick/tick-routing/file"
	"github.com/ltick/tick-routing/slash"
	"log"
	"net/http"
)

func Example() {
	router := routing.New(context.Background())

	router.AppendStartupHandler(
		// all these handlers are shared by every route
		access.Logger(log.Printf),
		slash.Remover(http.StatusMovedPermanently),
		fault.Recovery(log.Printf),
	)

	// serve RESTful APIs
	api := router.Group("/api", nil,nil)
	api.AppendStartupHandler(
		// these handlers are shared by the routes in the api group only
		content.TypeNegotiator(content.JSON, content.XML),
	)
	api.Get("/users", func(ctx context.Context, c *routing.Context) (context.Context, error) {
		return ctx, c.Write("user list")
	})
	api.Post("/users", func(ctx context.Context, c *routing.Context) (context.Context, error) {
		return ctx, c.Write("create a new user")
	})
	api.Put(`/users/<id:\d+>`, func(ctx context.Context, c *routing.Context) (context.Context, error) {
		return ctx, c.Write("update user " + c.Param("id"))
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
