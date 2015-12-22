package routing_test

import (
	"log"
	"net/http"
	"github.com/go-ozzo/ozzo-routing"
)

func Example() {
	r := routing.NewRouter()

	r.Use(
		routing.AccessLogger(log.Printf),
		routing.TrailingSlashRemover(http.StatusMovedPermanently),
	)

	r.Get("", func(c *routing.Context) {
		c.Write("Welcome, ozzo!")
	})
	r.Post("/login", func(c *routing.Context) {
		c.Write("Please login first")
	})
	r.Group("/admin", func(gr *routing.Router) {
		gr.Get("/posts", func(c *routing.Context) {
			c.Write("GET /admin/posts")
		})
		gr.Post("/posts", func(c *routing.Context) {
			c.Write("POST /admin/posts")
		})
	})

	r.Use(routing.NotFoundHandler())

	r.Error(routing.ErrorHandler(nil))

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
