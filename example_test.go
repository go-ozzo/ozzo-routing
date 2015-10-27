package routing_test

import (
	"os"
	"net/http"
	"github.com/go-ozzo/ozzo-routing"
)

func Example() {
	r := routing.NewRouter()

	r.Use(
		routing.AccessLogger(os.Stdout),
		routing.TrailingSlashRemover(http.StatusMovedPermanently),
	)

	r.Get("", func() string {
		return "Welcome, ozzo!"
	})
	r.Post("/login", func() string {
		return "Please login first"
	})
	r.Group("/admin", func(gr *routing.Router) {
		gr.Get("/posts", func() string {
			return "GET /admin/posts"
		})
		gr.Post("/posts", func() string {
			return "POST /admin/posts"
		})
	})

	r.Use(routing.NotFoundHandler())

	r.Error(routing.ErrorHandler(nil))

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
